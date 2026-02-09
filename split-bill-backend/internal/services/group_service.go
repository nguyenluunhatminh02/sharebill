package services

import (
	"context"
	"errors"
	"time"

	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/repository"
	"github.com/splitbill/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GroupService struct {
	groupRepo *repository.GroupRepository
	userRepo  *repository.UserRepository
}

func NewGroupService(groupRepo *repository.GroupRepository, userRepo *repository.UserRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// CreateGroup creates a new group with the creator as admin
func (s *GroupService) CreateGroup(ctx context.Context, creatorFirebaseUID string, req models.CreateGroupRequest) (*models.Group, error) {
	creator, err := s.userRepo.FindByFirebaseUID(ctx, creatorFirebaseUID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	group := &models.Group{
		Name:        req.Name,
		Description: req.Description,
		AvatarURL:   req.AvatarURL,
		CreatedBy:   creator.ID,
		Members: []models.GroupMember{
			{
				UserID:   creator.ID,
				Nickname: creator.DisplayName,
				Role:     models.RoleAdmin,
				JoinedAt: time.Now(),
			},
		},
		InviteCode: utils.GenerateInviteCode(),
		IsActive:   true,
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// GetGroup gets a group by ID, verifying the user is a member
func (s *GroupService) GetGroup(ctx context.Context, groupID string, firebaseUID string) (*models.Group, error) {
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("invalid group ID")
	}

	user, err := s.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	group, err := s.groupRepo.FindByID(ctx, objID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("group not found")
		}
		return nil, err
	}

	// Check membership
	isMember := false
	for _, m := range group.Members {
		if m.UserID == user.ID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, errors.New("you are not a member of this group")
	}

	return group, nil
}

// ListGroups lists all groups for a user
func (s *GroupService) ListGroups(ctx context.Context, firebaseUID string) ([]models.Group, error) {
	user, err := s.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return s.groupRepo.FindByMemberUserID(ctx, user.ID)
}

// UpdateGroup updates group details
func (s *GroupService) UpdateGroup(ctx context.Context, groupID string, firebaseUID string, req models.UpdateGroupRequest) (*models.Group, error) {
	group, err := s.GetGroup(ctx, groupID, firebaseUID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		group.Name = req.Name
	}
	if req.Description != "" {
		group.Description = req.Description
	}
	if req.AvatarURL != "" {
		group.AvatarURL = req.AvatarURL
	}

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// AddMember adds a member to a group
func (s *GroupService) AddMember(ctx context.Context, groupID string, firebaseUID string, req models.AddMemberRequest) error {
	group, err := s.GetGroup(ctx, groupID, firebaseUID)
	if err != nil {
		return err
	}

	newMemberID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	// Check if user exists
	newMember, err := s.userRepo.FindByID(ctx, newMemberID)
	if err != nil {
		return errors.New("user not found")
	}

	// Check if already a member
	for _, m := range group.Members {
		if m.UserID == newMemberID {
			return errors.New("user is already a member")
		}
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = newMember.DisplayName
	}

	member := models.GroupMember{
		UserID:   newMemberID,
		Nickname: nickname,
		Role:     models.RoleMember,
		JoinedAt: time.Now(),
	}

	return s.groupRepo.AddMember(ctx, group.ID, member)
}

// JoinGroup joins a group using an invite code
func (s *GroupService) JoinGroup(ctx context.Context, firebaseUID string, inviteCode string) (*models.Group, error) {
	user, err := s.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	group, err := s.groupRepo.FindByInviteCode(ctx, inviteCode)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("invalid invite code")
		}
		return nil, err
	}

	// Check if already a member
	for _, m := range group.Members {
		if m.UserID == user.ID {
			return group, nil // Already a member, just return group
		}
	}

	member := models.GroupMember{
		UserID:   user.ID,
		Nickname: user.DisplayName,
		Role:     models.RoleMember,
		JoinedAt: time.Now(),
	}

	if err := s.groupRepo.AddMember(ctx, group.ID, member); err != nil {
		return nil, err
	}

	// Refresh group data
	return s.groupRepo.FindByID(ctx, group.ID)
}

// RemoveMember removes a member from a group
func (s *GroupService) RemoveMember(ctx context.Context, groupID string, firebaseUID string, memberUserID string) error {
	group, err := s.GetGroup(ctx, groupID, firebaseUID)
	if err != nil {
		return err
	}

	user, err := s.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return errors.New("user not found")
	}

	// Check if requester is admin
	isAdmin := false
	for _, m := range group.Members {
		if m.UserID == user.ID && m.Role == models.RoleAdmin {
			isAdmin = true
			break
		}
	}

	memberObjID, err := primitive.ObjectIDFromHex(memberUserID)
	if err != nil {
		return errors.New("invalid member user ID")
	}

	// Can remove self, or admin can remove others
	if user.ID != memberObjID && !isAdmin {
		return errors.New("only admins can remove other members")
	}

	return s.groupRepo.RemoveMember(ctx, group.ID, memberObjID)
}

// DeleteGroup soft-deletes a group (admin only)
func (s *GroupService) DeleteGroup(ctx context.Context, groupID string, firebaseUID string) error {
	group, err := s.GetGroup(ctx, groupID, firebaseUID)
	if err != nil {
		return err
	}

	user, err := s.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return errors.New("user not found")
	}

	// Check if admin
	for _, m := range group.Members {
		if m.UserID == user.ID && m.Role == models.RoleAdmin {
			return s.groupRepo.Delete(ctx, group.ID)
		}
	}

	return errors.New("only admins can delete groups")
}

// GetGroupWithMemberDetails returns a group response with full member details
func (s *GroupService) GetGroupWithMemberDetails(ctx context.Context, group *models.Group) (*models.GroupResponse, error) {
	resp := group.ToResponse()

	// Fetch member details
	memberIDs := make([]primitive.ObjectID, len(group.Members))
	for i, m := range group.Members {
		memberIDs[i] = m.UserID
	}

	users, err := s.userRepo.FindByIDs(ctx, memberIDs)
	if err != nil {
		return &resp, nil // Return without details if fetch fails
	}

	userMap := make(map[string]*models.User)
	for i := range users {
		userMap[users[i].ID.Hex()] = &users[i]
	}

	for i, m := range resp.Members {
		if user, ok := userMap[m.UserID]; ok {
			resp.Members[i].DisplayName = user.DisplayName
			resp.Members[i].AvatarURL = user.AvatarURL
		}
	}

	return &resp, nil
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MemberRole represents the role of a member in a group
type MemberRole string

const (
	RoleAdmin  MemberRole = "admin"
	RoleMember MemberRole = "member"
)

// GroupMember represents a member within a group
type GroupMember struct {
	UserID   primitive.ObjectID `bson:"user_id" json:"user_id"`
	Nickname string             `bson:"nickname" json:"nickname"`
	Role     MemberRole         `bson:"role" json:"role"`
	JoinedAt time.Time          `bson:"joined_at" json:"joined_at"`
}

// Group represents a group of people splitting bills
type Group struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	AvatarURL   string             `bson:"avatar_url" json:"avatar_url"`
	CreatedBy   primitive.ObjectID `bson:"created_by" json:"created_by"`
	Members     []GroupMember      `bson:"members" json:"members"`
	InviteCode  string             `bson:"invite_code" json:"invite_code"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateGroupRequest is the request body for creating a group
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	AvatarURL   string `json:"avatar_url"`
}

// UpdateGroupRequest is the request body for updating a group
type UpdateGroupRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	AvatarURL   string `json:"avatar_url"`
}

// AddMemberRequest is the request body for adding a member to a group
type AddMemberRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Nickname string `json:"nickname"`
}

// JoinGroupRequest is the request body for joining a group by invite code
type JoinGroupRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// GroupResponse is the API response for a group
type GroupResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	AvatarURL   string                `json:"avatar_url"`
	CreatedBy   string                `json:"created_by"`
	Members     []GroupMemberResponse `json:"members"`
	InviteCode  string                `json:"invite_code"`
	IsActive    bool                  `json:"is_active"`
	CreatedAt   time.Time             `json:"created_at"`
}

// GroupMemberResponse is the API response for a group member
type GroupMemberResponse struct {
	UserID      string     `json:"user_id"`
	Nickname    string     `json:"nickname"`
	DisplayName string     `json:"display_name"`
	AvatarURL   string     `json:"avatar_url"`
	Role        MemberRole `json:"role"`
	JoinedAt    time.Time  `json:"joined_at"`
}

func (g *Group) ToResponse() GroupResponse {
	members := make([]GroupMemberResponse, len(g.Members))
	for i, m := range g.Members {
		members[i] = GroupMemberResponse{
			UserID:   m.UserID.Hex(),
			Nickname: m.Nickname,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		}
	}

	return GroupResponse{
		ID:          g.ID.Hex(),
		Name:        g.Name,
		Description: g.Description,
		AvatarURL:   g.AvatarURL,
		CreatedBy:   g.CreatedBy.Hex(),
		Members:     members,
		InviteCode:  g.InviteCode,
		IsActive:    g.IsActive,
		CreatedAt:   g.CreatedAt,
	}
}

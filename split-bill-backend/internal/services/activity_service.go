package services

import (
	"context"
	"fmt"

	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type ActivityService struct {
	activityRepo *repository.ActivityRepository
	userRepo     *repository.UserRepository
	groupRepo    *repository.GroupRepository
	logger       *zap.Logger
}

func NewActivityService(
	activityRepo *repository.ActivityRepository,
	userRepo *repository.UserRepository,
	groupRepo *repository.GroupRepository,
	logger *zap.Logger,
) *ActivityService {
	return &ActivityService{
		activityRepo: activityRepo,
		userRepo:     userRepo,
		groupRepo:    groupRepo,
		logger:       logger,
	}
}

// LogActivity creates a new activity record
func (s *ActivityService) LogActivity(ctx context.Context, groupID, userID primitive.ObjectID, actType models.ActivityType, title, detail string, amount float64, refID string) {
	activity := &models.Activity{
		GroupID: groupID,
		UserID:  userID,
		Type:    actType,
		Title:   title,
		Detail:  detail,
		Amount:  amount,
		RefID:   refID,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Error("Failed to log activity", zap.Error(err))
	}
}

// LogBillCreated logs a bill creation event
func (s *ActivityService) LogBillCreated(ctx context.Context, bill *models.Bill, creatorName string) {
	detail := fmt.Sprintf("%s đã tạo hóa đơn \"%s\" - %s", creatorName, bill.Title, formatVNDAmount(bill.TotalAmount))
	s.LogActivity(ctx, bill.GroupID, bill.PaidBy, models.ActivityBillCreated, "Hóa đơn mới", detail, bill.TotalAmount, bill.ID.Hex())
}

// LogPaymentSent logs a payment sent event
func (s *ActivityService) LogPaymentSent(ctx context.Context, tx *models.Transaction, fromName, toName string) {
	detail := fmt.Sprintf("%s đã gửi %s cho %s", fromName, formatVNDAmount(tx.Amount), toName)
	s.LogActivity(ctx, tx.GroupID, tx.FromUser, models.ActivityPaymentSent, "Thanh toán", detail, tx.Amount, tx.ID.Hex())
}

// LogPaymentConfirmed logs a payment confirmation event
func (s *ActivityService) LogPaymentConfirmed(ctx context.Context, tx *models.Transaction, confirmerName, senderName string) {
	detail := fmt.Sprintf("%s đã xác nhận thanh toán %s từ %s", confirmerName, formatVNDAmount(tx.Amount), senderName)
	s.LogActivity(ctx, tx.GroupID, tx.ToUser, models.ActivityPaymentConfirmed, "Xác nhận thanh toán", detail, tx.Amount, tx.ID.Hex())
}

// LogPaymentRejected logs a payment rejection event
func (s *ActivityService) LogPaymentRejected(ctx context.Context, tx *models.Transaction, rejecterName, senderName string) {
	detail := fmt.Sprintf("%s đã từ chối thanh toán %s từ %s", rejecterName, formatVNDAmount(tx.Amount), senderName)
	s.LogActivity(ctx, tx.GroupID, tx.ToUser, models.ActivityPaymentRejected, "Từ chối thanh toán", detail, tx.Amount, tx.ID.Hex())
}

// LogMemberJoined logs a member join event
func (s *ActivityService) LogMemberJoined(ctx context.Context, groupID, userID primitive.ObjectID, memberName, groupName string) {
	detail := fmt.Sprintf("%s đã tham gia nhóm \"%s\"", memberName, groupName)
	s.LogActivity(ctx, groupID, userID, models.ActivityMemberJoined, "Thành viên mới", detail, 0, "")
}

// GetGroupActivities gets activities for a specific group
func (s *ActivityService) GetGroupActivities(ctx context.Context, groupID string, limit int64) ([]models.ActivityResponse, error) {
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	activities, err := s.activityRepo.FindByGroupID(ctx, objID, limit)
	if err != nil {
		return nil, err
	}

	return s.enrichActivities(ctx, activities)
}

// GetUserActivities gets activities across all user's groups
func (s *ActivityService) GetUserActivities(ctx context.Context, userID primitive.ObjectID, limit int64) ([]models.ActivityResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}

	// Get all groups the user belongs to
	groups, err := s.groupRepo.FindByMemberUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return []models.ActivityResponse{}, nil
	}

	groupIDs := make([]primitive.ObjectID, len(groups))
	groupNameMap := make(map[string]string)
	for i, g := range groups {
		groupIDs[i] = g.ID
		groupNameMap[g.ID.Hex()] = g.Name
	}

	activities, err := s.activityRepo.FindByUserGroups(ctx, groupIDs, limit)
	if err != nil {
		return nil, err
	}

	responses, err := s.enrichActivities(ctx, activities)
	if err != nil {
		return nil, err
	}

	// Add group names
	for i := range responses {
		if name, ok := groupNameMap[responses[i].GroupID]; ok {
			responses[i].GroupName = name
		}
	}

	return responses, nil
}

// enrichActivities adds user details to activity responses
func (s *ActivityService) enrichActivities(ctx context.Context, activities []models.Activity) ([]models.ActivityResponse, error) {
	responses := make([]models.ActivityResponse, len(activities))
	userCache := make(map[string]*models.User)

	for i, activity := range activities {
		responses[i] = activity.ToResponse()

		userIDStr := activity.UserID.Hex()
		if user, ok := userCache[userIDStr]; ok {
			responses[i].UserName = user.DisplayName
			responses[i].UserAvatar = user.AvatarURL
		} else {
			user, err := s.userRepo.FindByID(ctx, activity.UserID)
			if err == nil {
				userCache[userIDStr] = user
				responses[i].UserName = user.DisplayName
				responses[i].UserAvatar = user.AvatarURL
			}
		}
	}

	return responses, nil
}

func formatVNDAmount(amount float64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("%.1ftr₫", amount/1000000)
	}
	if amount >= 1000 {
		return fmt.Sprintf("%.0fk₫", amount/1000)
	}
	return fmt.Sprintf("%.0f₫", amount)
}

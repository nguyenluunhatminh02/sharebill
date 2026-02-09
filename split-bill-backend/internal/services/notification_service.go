package services

import (
	"context"
	"fmt"

	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotifBillCreated       NotificationType = "bill_created"
	NotifBillSplit         NotificationType = "bill_split"
	NotifPaymentReceived   NotificationType = "payment_received"
	NotifPaymentConfirmed  NotificationType = "payment_confirmed"
	NotifGroupInvite       NotificationType = "group_invite"
	NotifMemberJoined      NotificationType = "member_joined"
	NotifSettlementReminder NotificationType = "settlement_reminder"
)

// Notification represents a notification to be sent
type Notification struct {
	Type    NotificationType       `json:"type"`
	Title   string                 `json:"title"`
	Body    string                 `json:"body"`
	Data    map[string]string      `json:"data"`
	UserIDs []primitive.ObjectID   `json:"user_ids"`
}

// NotificationService handles push notifications via FCM
type NotificationService struct {
	userRepo *repository.UserRepository
	logger   *zap.Logger
	// In production, add firebase.App and messaging.Client
	// fcmClient *messaging.Client
	enabled bool
}

// NewNotificationService creates a new notification service
func NewNotificationService(userRepo *repository.UserRepository, logger *zap.Logger) *NotificationService {
	return &NotificationService{
		userRepo: userRepo,
		logger:   logger,
		enabled:  false, // Disabled until FCM is configured
	}
}

// SendNotification sends a notification to specified users
func (s *NotificationService) SendNotification(ctx context.Context, notif *Notification) error {
	if !s.enabled {
		s.logger.Debug("Notification not sent (FCM disabled)",
			zap.String("type", string(notif.Type)),
			zap.String("title", notif.Title),
			zap.Int("recipients", len(notif.UserIDs)),
		)
		return nil
	}

	// In production:
	// 1. Get FCM tokens for user IDs from user repository
	// 2. Build FCM message
	// 3. Send via Firebase Cloud Messaging
	//
	// tokens, err := s.getUserFCMTokens(ctx, notif.UserIDs)
	// message := &messaging.MulticastMessage{
	//     Notification: &messaging.Notification{
	//         Title: notif.Title,
	//         Body:  notif.Body,
	//     },
	//     Data:   notif.Data,
	//     Tokens: tokens,
	// }
	// response, err := s.fcmClient.SendEachForMulticast(ctx, message)

	s.logger.Info("Notification sent",
		zap.String("type", string(notif.Type)),
		zap.Int("recipients", len(notif.UserIDs)),
	)

	return nil
}

// NotifyBillCreated notifies group members about a new bill
func (s *NotificationService) NotifyBillCreated(ctx context.Context, bill *models.Bill, creatorName string, groupName string, memberIDs []primitive.ObjectID) error {
	// Exclude the payer (bill creator) from notifications
	var recipients []primitive.ObjectID
	for _, id := range memberIDs {
		if id != bill.PaidBy {
			recipients = append(recipients, id)
		}
	}

	if len(recipients) == 0 {
		return nil
	}

	notif := &Notification{
		Type:    NotifBillCreated,
		Title:   fmt.Sprintf("New bill in %s", groupName),
		Body:    fmt.Sprintf("%s added \"%s\" - %s", creatorName, bill.Title, formatVND(bill.TotalAmount)),
		Data: map[string]string{
			"type":     string(NotifBillCreated),
			"bill_id":  bill.ID.Hex(),
			"group_id": bill.GroupID.Hex(),
		},
		UserIDs: recipients,
	}

	return s.SendNotification(ctx, notif)
}

// NotifyPaymentReceived notifies a user they received a payment
func (s *NotificationService) NotifyPaymentReceived(ctx context.Context, transaction *models.Transaction, fromUserName string) error {
	notif := &Notification{
		Type:    NotifPaymentReceived,
		Title:   "Payment Received",
		Body:    fmt.Sprintf("%s sent you %s", fromUserName, formatVND(transaction.Amount)),
		Data: map[string]string{
			"type":           string(NotifPaymentReceived),
			"transaction_id": transaction.ID.Hex(),
			"group_id":       transaction.GroupID.Hex(),
		},
		UserIDs: []primitive.ObjectID{transaction.ToUser},
	}

	return s.SendNotification(ctx, notif)
}

// NotifyPaymentConfirmed notifies a user their payment was confirmed
func (s *NotificationService) NotifyPaymentConfirmed(ctx context.Context, transaction *models.Transaction, confirmerName string) error {
	notif := &Notification{
		Type:    NotifPaymentConfirmed,
		Title:   "Payment Confirmed ✓",
		Body:    fmt.Sprintf("%s confirmed your payment of %s", confirmerName, formatVND(transaction.Amount)),
		Data: map[string]string{
			"type":           string(NotifPaymentConfirmed),
			"transaction_id": transaction.ID.Hex(),
			"group_id":       transaction.GroupID.Hex(),
		},
		UserIDs: []primitive.ObjectID{transaction.FromUser},
	}

	return s.SendNotification(ctx, notif)
}

// NotifyMemberJoined notifies group members when someone joins
func (s *NotificationService) NotifyMemberJoined(ctx context.Context, groupID primitive.ObjectID, groupName string, newMemberName string, memberIDs []primitive.ObjectID, excludeID primitive.ObjectID) error {
	var recipients []primitive.ObjectID
	for _, id := range memberIDs {
		if id != excludeID {
			recipients = append(recipients, id)
		}
	}

	if len(recipients) == 0 {
		return nil
	}

	notif := &Notification{
		Type:    NotifMemberJoined,
		Title:   groupName,
		Body:    fmt.Sprintf("%s joined the group", newMemberName),
		Data: map[string]string{
			"type":     string(NotifMemberJoined),
			"group_id": groupID.Hex(),
		},
		UserIDs: recipients,
	}

	return s.SendNotification(ctx, notif)
}

// NotifySettlementReminder sends a reminder about pending settlements
func (s *NotificationService) NotifySettlementReminder(ctx context.Context, userID primitive.ObjectID, groupName string, amount float64) error {
	notif := &Notification{
		Type:    NotifSettlementReminder,
		Title:   "Settlement Reminder 💰",
		Body:    fmt.Sprintf("You owe %s in %s", formatVND(amount), groupName),
		Data: map[string]string{
			"type": string(NotifSettlementReminder),
		},
		UserIDs: []primitive.ObjectID{userID},
	}

	return s.SendNotification(ctx, notif)
}

func formatVND(amount float64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("%.1fM₫", amount/1000000)
	}
	if amount >= 1000 {
		return fmt.Sprintf("%.0fK₫", amount/1000)
	}
	return fmt.Sprintf("%.0f₫", amount)
}

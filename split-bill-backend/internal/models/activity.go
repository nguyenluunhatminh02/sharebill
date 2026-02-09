package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityBillCreated       ActivityType = "bill_created"
	ActivityBillDeleted       ActivityType = "bill_deleted"
	ActivityBillUpdated       ActivityType = "bill_updated"
	ActivityMemberJoined      ActivityType = "member_joined"
	ActivityMemberLeft        ActivityType = "member_left"
	ActivityPaymentSent       ActivityType = "payment_sent"
	ActivityPaymentConfirmed  ActivityType = "payment_confirmed"
	ActivityPaymentRejected   ActivityType = "payment_rejected"
	ActivityGroupCreated      ActivityType = "group_created"
	ActivitySettlementCreated ActivityType = "settlement_created"
)

// Activity represents an activity event in a group
type Activity struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID   primitive.ObjectID `bson:"group_id" json:"group_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type      ActivityType       `bson:"type" json:"type"`
	Title     string             `bson:"title" json:"title"`
	Detail    string             `bson:"detail" json:"detail"`
	Amount    float64            `bson:"amount,omitempty" json:"amount,omitempty"`
	RefID     string             `bson:"ref_id,omitempty" json:"ref_id,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// ActivityResponse is the API response for an activity
type ActivityResponse struct {
	ID         string       `json:"id"`
	GroupID    string       `json:"group_id"`
	GroupName  string       `json:"group_name,omitempty"`
	UserID     string       `json:"user_id"`
	UserName   string       `json:"user_name"`
	UserAvatar string       `json:"user_avatar,omitempty"`
	Type       ActivityType `json:"type"`
	Title      string       `json:"title"`
	Detail     string       `json:"detail"`
	Amount     float64      `json:"amount,omitempty"`
	RefID      string       `json:"ref_id,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	TimeAgo    string       `json:"time_ago"`
}

func (a *Activity) ToResponse() ActivityResponse {
	return ActivityResponse{
		ID:        a.ID.Hex(),
		GroupID:   a.GroupID.Hex(),
		UserID:    a.UserID.Hex(),
		Type:      a.Type,
		Title:     a.Title,
		Detail:    a.Detail,
		Amount:    a.Amount,
		RefID:     a.RefID,
		CreatedAt: a.CreatedAt,
		TimeAgo:   timeAgo(a.CreatedAt),
	}
}

// timeAgo returns a human-readable time difference
func timeAgo(t time.Time) string {
	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "vừa xong"
	case diff < time.Hour:
		return fmt.Sprintf("%d phút trước", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d giờ trước", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		return fmt.Sprintf("%d ngày trước", int(diff.Hours()/24))
	case diff < 30*24*time.Hour:
		return fmt.Sprintf("%d tuần trước", int(diff.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%d tháng trước", int(diff.Hours()/(24*30)))
	}
}

// PaymentDeeplinkRequest is the request for generating a payment deeplink
type PaymentDeeplinkRequest struct {
	BankCode      string  `json:"bank_code" binding:"required"`
	AccountNumber string  `json:"account_number" binding:"required"`
	AccountName   string  `json:"account_name"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Note          string  `json:"note"`
}

// PaymentDeeplinkResponse contains deeplinks for various banking apps
type PaymentDeeplinkResponse struct {
	Amount    float64           `json:"amount"`
	Note      string            `json:"note"`
	Deeplinks []BankingDeeplink `json:"deeplinks"`
	VietQRURL string            `json:"vietqr_url"`
}

// BankingDeeplink represents a deeplink to a banking app
type BankingDeeplink struct {
	AppName  string `json:"app_name"`
	Scheme   string `json:"scheme"`
	Color    string `json:"color"`
	IconName string `json:"icon_name"`
}

// VietQRRequest is the request for generating a VietQR code
type VietQRRequest struct {
	BankID        string  `json:"bank_id" binding:"required"`
	AccountNumber string  `json:"account_no" binding:"required"`
	AccountName   string  `json:"account_name"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Description   string  `json:"description"`
	Template      string  `json:"template"`
}

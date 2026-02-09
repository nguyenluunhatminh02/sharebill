package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionPayment    TransactionType = "payment"
	TransactionSettlement TransactionType = "settlement"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "pending"
	TransactionConfirmed TransactionStatus = "confirmed"
	TransactionRejected  TransactionStatus = "rejected"
)

// Transaction represents a payment between users
type Transaction struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID         primitive.ObjectID `bson:"group_id" json:"group_id"`
	FromUser        primitive.ObjectID `bson:"from_user" json:"from_user"`
	ToUser          primitive.ObjectID `bson:"to_user" json:"to_user"`
	Amount          float64            `bson:"amount" json:"amount"`
	Currency        string             `bson:"currency" json:"currency"`
	BillID          primitive.ObjectID `bson:"bill_id,omitempty" json:"bill_id,omitempty"`
	Type            TransactionType    `bson:"type" json:"type"`
	Status          TransactionStatus  `bson:"status" json:"status"`
	PaymentMethod   string             `bson:"payment_method" json:"payment_method"`
	PaymentProofURL string             `bson:"payment_proof_url" json:"payment_proof_url"`
	Note            string             `bson:"note" json:"note"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	ConfirmedAt     *time.Time         `bson:"confirmed_at,omitempty" json:"confirmed_at,omitempty"`
}

// CreateTransactionRequest is the request body for creating a transaction
type CreateTransactionRequest struct {
	GroupID         string  `json:"group_id" binding:"required"`
	ToUser          string  `json:"to_user" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Currency        string  `json:"currency" binding:"required"`
	BillID          string  `json:"bill_id"`
	PaymentMethod   string  `json:"payment_method"`
	PaymentProofURL string  `json:"payment_proof_url"`
	Note            string  `json:"note" binding:"max=500"`
}

// TransactionResponse is the API response for a transaction
type TransactionResponse struct {
	ID              string            `json:"id"`
	GroupID         string            `json:"group_id"`
	FromUser        string            `json:"from_user"`
	FromUserName    string            `json:"from_user_name"`
	ToUser          string            `json:"to_user"`
	ToUserName      string            `json:"to_user_name"`
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	BillID          string            `json:"bill_id,omitempty"`
	Type            TransactionType   `json:"type"`
	Status          TransactionStatus `json:"status"`
	PaymentMethod   string            `json:"payment_method"`
	PaymentProofURL string            `json:"payment_proof_url"`
	Note            string            `json:"note"`
	CreatedAt       time.Time         `json:"created_at"`
	ConfirmedAt     *time.Time        `json:"confirmed_at,omitempty"`
}

func (t *Transaction) ToResponse() TransactionResponse {
	billID := ""
	if !t.BillID.IsZero() {
		billID = t.BillID.Hex()
	}
	return TransactionResponse{
		ID:              t.ID.Hex(),
		GroupID:         t.GroupID.Hex(),
		FromUser:        t.FromUser.Hex(),
		ToUser:          t.ToUser.Hex(),
		Amount:          t.Amount,
		Currency:        t.Currency,
		BillID:          billID,
		Type:            t.Type,
		Status:          t.Status,
		PaymentMethod:   t.PaymentMethod,
		PaymentProofURL: t.PaymentProofURL,
		Note:            t.Note,
		CreatedAt:       t.CreatedAt,
		ConfirmedAt:     t.ConfirmedAt,
	}
}

// Settlement represents an optimized payment suggestion
type Settlement struct {
	FromUserID   string  `json:"from_user_id"`
	FromUserName string  `json:"from_user_name"`
	ToUserID     string  `json:"to_user_id"`
	ToUserName   string  `json:"to_user_name"`
	Amount       float64 `json:"amount"`
}

// BalanceResponse represents the balance info for a user in a group
type BalanceResponse struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Balance     float64 `json:"balance"` // positive = owed money, negative = owes money
}

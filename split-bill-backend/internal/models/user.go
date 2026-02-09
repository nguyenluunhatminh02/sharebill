package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BankAccount represents a user's bank account info
type BankAccount struct {
	BankCode      string `bson:"bank_code" json:"bank_code"`
	AccountNumber string `bson:"account_number" json:"account_number"`
	AccountName   string `bson:"account_name" json:"account_name"`
}

// User represents a user in the system
type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirebaseUID      string             `bson:"firebase_uid" json:"firebase_uid"`
	Phone            string             `bson:"phone" json:"phone"`
	DisplayName      string             `bson:"display_name" json:"display_name"`
	AvatarURL        string             `bson:"avatar_url" json:"avatar_url"`
	BankAccounts     []BankAccount      `bson:"bank_accounts" json:"bank_accounts"`
	PreferredPayment string             `bson:"preferred_payment" json:"preferred_payment"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateUserRequest is the request body for creating/updating a user
type CreateUserRequest struct {
	Phone       string `json:"phone" binding:"required"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=50"`
	AvatarURL   string `json:"avatar_url"`
}

// UpdateUserRequest is the request body for updating a user profile
type UpdateUserRequest struct {
	DisplayName      string        `json:"display_name" binding:"omitempty,min=2,max=50"`
	AvatarURL        string        `json:"avatar_url"`
	BankAccounts     []BankAccount `json:"bank_accounts"`
	PreferredPayment string        `json:"preferred_payment"`
}

// UserResponse is the response for user info
type UserResponse struct {
	ID               string        `json:"id"`
	Phone            string        `json:"phone"`
	DisplayName      string        `json:"display_name"`
	AvatarURL        string        `json:"avatar_url"`
	BankAccounts     []BankAccount `json:"bank_accounts"`
	PreferredPayment string        `json:"preferred_payment"`
	CreatedAt        time.Time     `json:"created_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:               u.ID.Hex(),
		Phone:            u.Phone,
		DisplayName:      u.DisplayName,
		AvatarURL:        u.AvatarURL,
		BankAccounts:     u.BankAccounts,
		PreferredPayment: u.PreferredPayment,
		CreatedAt:        u.CreatedAt,
	}
}

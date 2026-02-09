package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SplitType represents how a bill is split
type SplitType string

const (
	SplitEqual      SplitType = "equal"
	SplitByItem     SplitType = "by_item"
	SplitByPercent  SplitType = "by_percentage"
	SplitByAmount   SplitType = "by_amount"
)

// BillStatus represents the status of a bill
type BillStatus string

const (
	BillPending   BillStatus = "pending"
	BillSettled   BillStatus = "settled"
	BillCancelled BillStatus = "cancelled"
)

// BillItem represents a single item on a bill
type BillItem struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name       string               `bson:"name" json:"name"`
	Quantity   int                  `bson:"quantity" json:"quantity"`
	UnitPrice  float64              `bson:"unit_price" json:"unit_price"`
	TotalPrice float64              `bson:"total_price" json:"total_price"`
	AssignedTo []primitive.ObjectID `bson:"assigned_to" json:"assigned_to"`
}

// ExtraCharges represents additional charges on a bill
type ExtraCharges struct {
	Tax           float64 `bson:"tax" json:"tax"`
	ServiceCharge float64 `bson:"service_charge" json:"service_charge"`
	Tip           float64 `bson:"tip" json:"tip"`
	Discount      float64 `bson:"discount" json:"discount"`
}

// BillSplit represents how much a user owes for a bill
type BillSplit struct {
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`
	Amount float64            `bson:"amount" json:"amount"`
	IsPaid bool               `bson:"is_paid" json:"is_paid"`
	PaidAt *time.Time         `bson:"paid_at,omitempty" json:"paid_at,omitempty"`
}

// Bill represents a bill/expense in a group
type Bill struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID         primitive.ObjectID `bson:"group_id" json:"group_id"`
	Title           string             `bson:"title" json:"title"`
	Description     string             `bson:"description" json:"description"`
	Category        string             `bson:"category" json:"category"`
	ReceiptImageURL string             `bson:"receipt_image_url" json:"receipt_image_url"`
	TotalAmount     float64            `bson:"total_amount" json:"total_amount"`
	Currency        string             `bson:"currency" json:"currency"`
	PaidBy          primitive.ObjectID `bson:"paid_by" json:"paid_by"`
	SplitType       SplitType          `bson:"split_type" json:"split_type"`
	Items           []BillItem         `bson:"items" json:"items"`
	ExtraCharges    ExtraCharges       `bson:"extra_charges" json:"extra_charges"`
	Splits          []BillSplit        `bson:"splits" json:"splits"`
	Status          BillStatus         `bson:"status" json:"status"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateBillRequest is the request body for creating a bill
type CreateBillRequest struct {
	Title           string              `json:"title" binding:"required,min=2,max=200"`
	Description     string              `json:"description" binding:"max=500"`
	Category        string              `json:"category"`
	ReceiptImageURL string              `json:"receipt_image_url"`
	TotalAmount     float64             `json:"total_amount" binding:"required,gt=0"`
	Currency        string              `json:"currency" binding:"required"`
	PaidBy          string              `json:"paid_by" binding:"required"`
	SplitType       SplitType           `json:"split_type" binding:"required"`
	Items           []CreateBillItemReq `json:"items"`
	ExtraCharges    ExtraCharges        `json:"extra_charges"`
	SplitAmong      []string            `json:"split_among"` // user IDs for equal split
}

// CreateBillItemReq is the request for a bill item
type CreateBillItemReq struct {
	Name       string   `json:"name" binding:"required"`
	Quantity   int      `json:"quantity" binding:"required,gt=0"`
	UnitPrice  float64  `json:"unit_price" binding:"required,gt=0"`
	TotalPrice float64  `json:"total_price"`
	AssignedTo []string `json:"assigned_to"`
}

// UpdateBillRequest is the request body for updating a bill
type UpdateBillRequest struct {
	Title        string       `json:"title" binding:"omitempty,min=2,max=200"`
	Description  string       `json:"description" binding:"max=500"`
	Category     string       `json:"category"`
	TotalAmount  *float64     `json:"total_amount" binding:"omitempty,gt=0"`
	ExtraCharges ExtraCharges `json:"extra_charges"`
	Status       BillStatus   `json:"status"`
}

// BillResponse is the API response for a bill
type BillResponse struct {
	ID              string              `json:"id"`
	GroupID         string              `json:"group_id"`
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	Category        string              `json:"category"`
	ReceiptImageURL string              `json:"receipt_image_url"`
	TotalAmount     float64             `json:"total_amount"`
	Currency        string              `json:"currency"`
	PaidBy          string              `json:"paid_by"`
	PaidByName      string              `json:"paid_by_name"`
	SplitType       SplitType           `json:"split_type"`
	Items           []BillItemResponse  `json:"items"`
	ExtraCharges    ExtraCharges        `json:"extra_charges"`
	Splits          []BillSplitResponse `json:"splits"`
	Status          BillStatus          `json:"status"`
	CreatedAt       time.Time           `json:"created_at"`
}

// BillItemResponse is the API response for a bill item
type BillItemResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Quantity   int      `json:"quantity"`
	UnitPrice  float64  `json:"unit_price"`
	TotalPrice float64  `json:"total_price"`
	AssignedTo []string `json:"assigned_to"`
}

// BillSplitResponse is the API response for a bill split
type BillSplitResponse struct {
	UserID      string     `json:"user_id"`
	DisplayName string     `json:"display_name"`
	Amount      float64    `json:"amount"`
	IsPaid      bool       `json:"is_paid"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
}

func (b *Bill) ToResponse() BillResponse {
	items := make([]BillItemResponse, len(b.Items))
	for i, item := range b.Items {
		assignedTo := make([]string, len(item.AssignedTo))
		for j, id := range item.AssignedTo {
			assignedTo[j] = id.Hex()
		}
		items[i] = BillItemResponse{
			ID:         item.ID.Hex(),
			Name:       item.Name,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
			AssignedTo: assignedTo,
		}
	}

	splits := make([]BillSplitResponse, len(b.Splits))
	for i, split := range b.Splits {
		splits[i] = BillSplitResponse{
			UserID: split.UserID.Hex(),
			Amount: split.Amount,
			IsPaid: split.IsPaid,
			PaidAt: split.PaidAt,
		}
	}

	return BillResponse{
		ID:              b.ID.Hex(),
		GroupID:         b.GroupID.Hex(),
		Title:           b.Title,
		Description:     b.Description,
		Category:        b.Category,
		ReceiptImageURL: b.ReceiptImageURL,
		TotalAmount:     b.TotalAmount,
		Currency:        b.Currency,
		PaidBy:          b.PaidBy.Hex(),
		SplitType:       b.SplitType,
		Items:           items,
		ExtraCharges:    b.ExtraCharges,
		Splits:          splits,
		Status:          b.Status,
		CreatedAt:       b.CreatedAt,
	}
}

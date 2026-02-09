package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OCRResult represents a scanned receipt OCR result
type OCRResult struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BillID            primitive.ObjectID `json:"bill_id,omitempty" bson:"bill_id,omitempty"`
	GroupID           primitive.ObjectID `json:"group_id" bson:"group_id"`
	UploadedBy        primitive.ObjectID `json:"uploaded_by" bson:"uploaded_by"`
	ImageURL          string             `json:"image_url" bson:"image_url"`
	RawText           string             `json:"raw_text" bson:"raw_text"`
	ParsedItems       []ParsedItem       `json:"parsed_items" bson:"parsed_items"`
	ParsedTotal       float64            `json:"parsed_total" bson:"parsed_total"`
	ParsedTax         float64            `json:"parsed_tax" bson:"parsed_tax"`
	ParsedServiceFee  float64            `json:"parsed_service_fee" bson:"parsed_service_fee"`
	ParsedDiscount    float64            `json:"parsed_discount" bson:"parsed_discount"`
	ConfidenceScore   float64            `json:"confidence_score" bson:"confidence_score"`
	ProcessingTimeMs  int64              `json:"processing_time_ms" bson:"processing_time_ms"`
	Status            OCRStatus          `json:"status" bson:"status"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	ConfirmedAt       *time.Time         `json:"confirmed_at,omitempty" bson:"confirmed_at,omitempty"`
}

// ParsedItem represents a single item parsed from a receipt
type ParsedItem struct {
	Name       string  `json:"name" bson:"name"`
	Quantity   int     `json:"quantity" bson:"quantity"`
	UnitPrice  float64 `json:"unit_price" bson:"unit_price"`
	TotalPrice float64 `json:"total_price" bson:"total_price"`
	Confidence float64 `json:"confidence" bson:"confidence"` // 0.0 - 1.0
}

// OCRStatus represents the status of an OCR scan
type OCRStatus string

const (
	OCRStatusProcessing OCRStatus = "processing"
	OCRStatusCompleted  OCRStatus = "completed"
	OCRStatusFailed     OCRStatus = "failed"
	OCRStatusConfirmed  OCRStatus = "confirmed"
)

// ScanReceiptRequest represents the request to scan a receipt
type ScanReceiptRequest struct {
	GroupID  string `json:"group_id" binding:"required"`
	ImageURL string `json:"image_url" binding:"required"`
}

// ScanReceiptFromBase64Request for direct base64 image upload
type ScanReceiptFromBase64Request struct {
	GroupID     string `json:"group_id" binding:"required"`
	ImageBase64 string `json:"image_base64" binding:"required"`
	FileName    string `json:"file_name"`
}

// ConfirmOCRRequest represents the request to confirm OCR results
type ConfirmOCRRequest struct {
	Title       string       `json:"title" binding:"required"`
	Items       []ParsedItem `json:"items" binding:"required"`
	Total       float64      `json:"total" binding:"required"`
	Tax         float64      `json:"tax"`
	ServiceFee  float64      `json:"service_fee"`
	Discount    float64      `json:"discount"`
	PaidBy      string       `json:"paid_by" binding:"required"`
	SplitType   string       `json:"split_type" binding:"required"`
	SplitAmong  []string     `json:"split_among"` // user IDs for equal split
}

// OCRResultResponse represents the response for OCR results
type OCRResultResponse struct {
	ID               string       `json:"id"`
	ImageURL         string       `json:"image_url"`
	RawText          string       `json:"raw_text"`
	ParsedItems      []ParsedItem `json:"parsed_items"`
	ParsedTotal      float64      `json:"parsed_total"`
	ParsedTax        float64      `json:"parsed_tax"`
	ParsedServiceFee float64      `json:"parsed_service_fee"`
	ParsedDiscount   float64      `json:"parsed_discount"`
	ConfidenceScore  float64      `json:"confidence_score"`
	ProcessingTimeMs int64        `json:"processing_time_ms"`
	Status           OCRStatus    `json:"status"`
	CreatedAt        time.Time    `json:"created_at"`
}

// ImageUploadResponse represents the response after image upload
type ImageUploadResponse struct {
	URL      string `json:"url"`
	FileName string `json:"file_name"`
}

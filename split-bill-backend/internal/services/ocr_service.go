package services

import (
	"context"
	"fmt"
	"time"

	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/repository"
	"github.com/splitbill/backend/internal/utils"
	"github.com/splitbill/backend/pkg/visionapi"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type OCRService struct {
	ocrRepo   *repository.OCRRepository
	billRepo  *repository.BillRepository
	groupRepo *repository.GroupRepository
	vision    *visionapi.Client
	parser    *utils.ReceiptParser
	logger    *zap.Logger
}

func NewOCRService(
	ocrRepo *repository.OCRRepository,
	billRepo *repository.BillRepository,
	groupRepo *repository.GroupRepository,
	vision *visionapi.Client,
	logger *zap.Logger,
) *OCRService {
	return &OCRService{
		ocrRepo:   ocrRepo,
		billRepo:  billRepo,
		groupRepo: groupRepo,
		vision:    vision,
		parser:    utils.NewReceiptParser(),
		logger:    logger,
	}
}

// ScanReceipt processes a receipt image through OCR
func (s *OCRService) ScanReceipt(ctx context.Context, userID, groupID primitive.ObjectID, imageURL string) (*models.OCRResult, error) {
	// Verify user is a member of the group
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("user is not a member of this group")
	}

	// Create OCR result record with processing status
	ocrResult := &models.OCRResult{
		GroupID:    groupID,
		UploadedBy: userID,
		ImageURL:   imageURL,
		Status:     models.OCRStatusProcessing,
	}

	if err := s.ocrRepo.Create(ctx, ocrResult); err != nil {
		return nil, fmt.Errorf("failed to create OCR record: %w", err)
	}

	// Process OCR asynchronously-ish (in same request for simplicity)
	startTime := time.Now()

	// Call Vision API
	rawText, err := s.vision.DetectText(ctx, imageURL)
	if err != nil {
		ocrResult.Status = models.OCRStatusFailed
		ocrResult.RawText = fmt.Sprintf("OCR Error: %s", err.Error())
		_ = s.ocrRepo.Update(ctx, ocrResult)
		return nil, fmt.Errorf("OCR detection failed: %w", err)
	}

	// Parse the raw text
	parseResult := s.parser.Parse(rawText)

	// Update OCR result
	processingTime := time.Since(startTime).Milliseconds()
	ocrResult.RawText = rawText
	ocrResult.ParsedItems = parseResult.Items
	ocrResult.ParsedTotal = parseResult.Total
	ocrResult.ParsedTax = parseResult.Tax
	ocrResult.ParsedServiceFee = parseResult.ServiceFee
	ocrResult.ParsedDiscount = parseResult.Discount
	ocrResult.ConfidenceScore = parseResult.Confidence
	ocrResult.ProcessingTimeMs = processingTime
	ocrResult.Status = models.OCRStatusCompleted

	if err := s.ocrRepo.Update(ctx, ocrResult); err != nil {
		s.logger.Error("Failed to update OCR result", zap.Error(err))
	}

	s.logger.Info("Receipt scanned successfully",
		zap.String("ocr_id", ocrResult.ID.Hex()),
		zap.Int("items_found", len(parseResult.Items)),
		zap.Float64("total", parseResult.Total),
		zap.Float64("confidence", parseResult.Confidence),
		zap.Int64("processing_ms", processingTime),
	)

	return ocrResult, nil
}

// ScanReceiptBase64 processes a base64-encoded receipt image
func (s *OCRService) ScanReceiptBase64(ctx context.Context, userID, groupID primitive.ObjectID, base64Image string) (*models.OCRResult, error) {
	// For base64, we pass it directly to the vision client
	return s.ScanReceipt(ctx, userID, groupID, base64Image)
}

// GetOCRResult retrieves an OCR result by ID
func (s *OCRService) GetOCRResult(ctx context.Context, ocrID primitive.ObjectID) (*models.OCRResult, error) {
	return s.ocrRepo.FindByID(ctx, ocrID)
}

// ConfirmOCR confirms OCR results and creates a bill from them
func (s *OCRService) ConfirmOCR(ctx context.Context, ocrID primitive.ObjectID, userID primitive.ObjectID, req *models.ConfirmOCRRequest) (*models.Bill, error) {
	// Get OCR result
	ocrResult, err := s.ocrRepo.FindByID(ctx, ocrID)
	if err != nil {
		return nil, fmt.Errorf("OCR result not found: %w", err)
	}

	if ocrResult.Status == models.OCRStatusConfirmed {
		return nil, fmt.Errorf("OCR result already confirmed")
	}

	// Parse paid_by
	paidByID, err := primitive.ObjectIDFromHex(req.PaidBy)
	if err != nil {
		return nil, fmt.Errorf("invalid paid_by user ID: %w", err)
	}

	// Build bill items from confirmed parsed items
	var billItems []models.BillItem
	for _, item := range req.Items {
		billItems = append(billItems, models.BillItem{
			Name:       item.Name,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		})
	}

	// Create bill
	bill := &models.Bill{
		GroupID:    ocrResult.GroupID,
		Title:     req.Title,
		PaidBy:    paidByID,
		TotalAmount: req.Total,
		Currency:    "VND",
		SplitType:   models.SplitType(req.SplitType),
		Items:       billItems,
		ExtraCharges: models.ExtraCharges{
			Tax:           req.Tax,
			ServiceCharge: req.ServiceFee,
			Discount:      req.Discount,
		},
		ReceiptImageURL: ocrResult.ImageURL,
		Status:          models.BillPending,
	}

	// Calculate splits based on split type
	if req.SplitType == string(models.SplitEqual) && len(req.SplitAmong) > 0 {
		// Equal split among specified users
		splitAmount := req.Total / float64(len(req.SplitAmong))
		for _, userIDStr := range req.SplitAmong {
			uid, parseErr := primitive.ObjectIDFromHex(userIDStr)
			if parseErr != nil {
				continue
			}
			bill.Splits = append(bill.Splits, models.BillSplit{
				UserID: uid,
				Amount: splitAmount,
			})
		}
	}

	if err := s.billRepo.Create(ctx, bill); err != nil {
		return nil, fmt.Errorf("failed to create bill: %w", err)
	}

	// Update OCR result with bill reference
	if err := s.ocrRepo.SetBillID(ctx, ocrID, bill.ID); err != nil {
		s.logger.Error("Failed to set bill ID on OCR result", zap.Error(err))
	}
	if err := s.ocrRepo.UpdateStatus(ctx, ocrID, models.OCRStatusConfirmed); err != nil {
		s.logger.Error("Failed to update OCR status", zap.Error(err))
	}

	s.logger.Info("OCR confirmed and bill created",
		zap.String("ocr_id", ocrID.Hex()),
		zap.String("bill_id", bill.ID.Hex()),
	)

	return bill, nil
}

// GetGroupScanHistory returns OCR scan history for a group
func (s *OCRService) GetGroupScanHistory(ctx context.Context, groupID primitive.ObjectID, limit int64) ([]models.OCRResult, error) {
	return s.ocrRepo.FindByGroupID(ctx, groupID, limit)
}

// GetPendingScans returns pending (unconfirmed) OCR scans for a user
func (s *OCRService) GetPendingScans(ctx context.Context, userID primitive.ObjectID) ([]models.OCRResult, error) {
	return s.ocrRepo.FindPendingByUser(ctx, userID)
}

package handlers

import (
	"net/http"

	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/services"
	"github.com/splitbill/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OCRHandler struct {
	ocrService *services.OCRService
}

func NewOCRHandler(ocrService *services.OCRService) *OCRHandler {
	return &OCRHandler{ocrService: ocrService}
}

// ScanReceipt handles receipt scanning via image URL
func (h *OCRHandler) ScanReceipt(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.ScanReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	groupID, err := primitive.ObjectIDFromHex(req.GroupID)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid group ID")
		return
	}

	result, err := h.ocrService.ScanReceipt(c.Request.Context(), userObjID, groupID, req.ImageURL)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Receipt scanned successfully", toOCRResponse(result))
}

// ScanReceiptBase64 handles receipt scanning via base64-encoded image
func (h *OCRHandler) ScanReceiptBase64(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.ScanReceiptFromBase64Request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	groupID, err := primitive.ObjectIDFromHex(req.GroupID)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid group ID")
		return
	}

	result, err := h.ocrService.ScanReceiptBase64(c.Request.Context(), userObjID, groupID, req.ImageBase64)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Receipt scanned successfully", toOCRResponse(result))
}

// GetOCRResult retrieves an OCR scan result
func (h *OCRHandler) GetOCRResult(c *gin.Context) {
	ocrIDStr := c.Param("id")
	ocrID, err := primitive.ObjectIDFromHex(ocrIDStr)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid OCR result ID")
		return
	}

	result, err := h.ocrService.GetOCRResult(c.Request.Context(), ocrID)
	if err != nil {
		utils.RespondError(c, http.StatusNotFound, "OCR result not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "OCR result retrieved", toOCRResponse(result))
}

// ConfirmOCR confirms parsed OCR results and creates a bill
func (h *OCRHandler) ConfirmOCR(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	ocrIDStr := c.Param("id")
	ocrID, err := primitive.ObjectIDFromHex(ocrIDStr)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid OCR result ID")
		return
	}

	var req models.ConfirmOCRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	bill, err := h.ocrService.ConfirmOCR(c.Request.Context(), ocrID, userObjID, &req)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, "Bill created from OCR scan", bill)
}

// GetPendingScans returns pending OCR scans for the current user
func (h *OCRHandler) GetPendingScans(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	results, err := h.ocrService.GetPendingScans(c.Request.Context(), userObjID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	var responses []models.OCRResultResponse
	for _, r := range results {
		responses = append(responses, *toOCRResponse(&r))
	}

	utils.RespondSuccess(c, http.StatusOK, "Pending scans retrieved", responses)
}

func toOCRResponse(result *models.OCRResult) *models.OCRResultResponse {
	return &models.OCRResultResponse{
		ID:               result.ID.Hex(),
		ImageURL:         result.ImageURL,
		RawText:          result.RawText,
		ParsedItems:      result.ParsedItems,
		ParsedTotal:      result.ParsedTotal,
		ParsedTax:        result.ParsedTax,
		ParsedServiceFee: result.ParsedServiceFee,
		ParsedDiscount:   result.ParsedDiscount,
		ConfidenceScore:  result.ConfidenceScore,
		ProcessingTimeMs: result.ProcessingTimeMs,
		Status:           result.Status,
		CreatedAt:        result.CreatedAt,
	}
}

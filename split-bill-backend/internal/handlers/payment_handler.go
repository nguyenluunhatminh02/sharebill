package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/repository"
	"github.com/splitbill/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentHandler struct {
	userRepo *repository.UserRepository
}

func NewPaymentHandler(userRepo *repository.UserRepository) *PaymentHandler {
	return &PaymentHandler{
		userRepo: userRepo,
	}
}

// GenerateDeeplink generates banking app deeplinks for payment
// POST /api/v1/payment/deeplink
func (h *PaymentHandler) GenerateDeeplink(c *gin.Context) {
	var req models.PaymentDeeplinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	note := req.Note
	if note == "" {
		note = "Split Bill Payment"
	}

	encodedNote := url.QueryEscape(note)
	amountStr := fmt.Sprintf("%.0f", req.Amount)

	// Generate deeplinks for Vietnamese banking apps
	deeplinks := []models.BankingDeeplink{
		{
			AppName:  "Momo",
			Scheme:   fmt.Sprintf("momo://transfer?phone=%s&amount=%s&note=%s", req.AccountNumber, amountStr, encodedNote),
			Color:    "#A6327F",
			IconName: "wallet",
		},
		{
			AppName:  "ZaloPay",
			Scheme:   fmt.Sprintf("zalopay://transfer?phone=%s&amount=%s&note=%s", req.AccountNumber, amountStr, encodedNote),
			Color:    "#008FE5",
			IconName: "wallet",
		},
		{
			AppName:  "VNPay QR",
			Scheme:   fmt.Sprintf("vnpayqr://pay?amount=%s&desc=%s", amountStr, encodedNote),
			Color:    "#1A3C7B",
			IconName: "credit-card",
		},
		{
			AppName:  "Vietcombank",
			Scheme:   fmt.Sprintf("vcbdigibank://transfer?account=%s&amount=%s&note=%s", req.AccountNumber, amountStr, encodedNote),
			Color:    "#00573F",
			IconName: "bank",
		},
		{
			AppName:  "Techcombank",
			Scheme:   fmt.Sprintf("techcombank://transfer?account=%s&amount=%s&note=%s", req.AccountNumber, amountStr, encodedNote),
			Color:    "#E31937",
			IconName: "bank",
		},
		{
			AppName:  "VPBank",
			Scheme:   fmt.Sprintf("vpbank://transfer?account=%s&amount=%s&note=%s", req.AccountNumber, amountStr, encodedNote),
			Color:    "#00A650",
			IconName: "bank",
		},
		{
			AppName:  "MB Bank",
			Scheme:   fmt.Sprintf("mbbank://transfer?account=%s&amount=%s&note=%s", req.AccountNumber, amountStr, encodedNote),
			Color:    "#005BAA",
			IconName: "bank",
		},
		{
			AppName:  "ACB",
			Scheme:   fmt.Sprintf("acb://transfer?account=%s&amount=%s&note=%s", req.AccountNumber, amountStr, encodedNote),
			Color:    "#1C4587",
			IconName: "bank",
		},
	}

	// Generate VietQR URL using vietqr.io API
	vietQRURL := generateVietQRURL(req.BankCode, req.AccountNumber, req.AccountName, req.Amount, note)

	response := models.PaymentDeeplinkResponse{
		Amount:    req.Amount,
		Note:      note,
		Deeplinks: deeplinks,
		VietQRURL: vietQRURL,
	}

	utils.RespondSuccess(c, http.StatusOK, "Payment deeplinks generated", response)
}

// GenerateVietQR generates a VietQR code image URL
// POST /api/v1/payment/vietqr
func (h *PaymentHandler) GenerateVietQR(c *gin.Context) {
	var req models.VietQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	template := req.Template
	if template == "" {
		template = "compact2"
	}

	description := req.Description
	if description == "" {
		description = "Split Bill Payment"
	}

	qrURL := fmt.Sprintf(
		"https://img.vietqr.io/image/%s-%s-%s.png?amount=%.0f&addInfo=%s&accountName=%s",
		req.BankID,
		req.AccountNumber,
		template,
		req.Amount,
		url.QueryEscape(description),
		url.QueryEscape(req.AccountName),
	)

	utils.RespondSuccess(c, http.StatusOK, "VietQR generated", gin.H{
		"qr_url":         qrURL,
		"bank_id":        req.BankID,
		"account_number": req.AccountNumber,
		"account_name":   req.AccountName,
		"amount":         req.Amount,
		"description":    description,
	})
}

// GetPaymentInfo gets payment info for a user (bank accounts)
// GET /api/v1/payment/user/:userId
func (h *PaymentHandler) GetUserPaymentInfo(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		utils.RespondNotFound(c, "User not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "User payment info", gin.H{
		"user_id":           user.ID.Hex(),
		"display_name":      user.DisplayName,
		"bank_accounts":     user.BankAccounts,
		"preferred_payment": user.PreferredPayment,
	})
}

// GetSupportedBanks returns the list of supported Vietnamese banks
// GET /api/v1/payment/banks
func (h *PaymentHandler) GetSupportedBanks(c *gin.Context) {
	banks := []gin.H{
		{"id": "VCB", "name": "Vietcombank", "code": "970436", "short_name": "VCB", "logo": "https://img.vietqr.io/img/VCB.png", "color": "#00573F"},
		{"id": "TCB", "name": "Techcombank", "code": "970407", "short_name": "TCB", "logo": "https://img.vietqr.io/img/TCB.png", "color": "#E31937"},
		{"id": "VPB", "name": "VPBank", "code": "970432", "short_name": "VPB", "logo": "https://img.vietqr.io/img/VPB.png", "color": "#00A650"},
		{"id": "MB", "name": "MB Bank", "code": "970422", "short_name": "MB", "logo": "https://img.vietqr.io/img/MB.png", "color": "#005BAA"},
		{"id": "ACB", "name": "ACB", "code": "970416", "short_name": "ACB", "logo": "https://img.vietqr.io/img/ACB.png", "color": "#1C4587"},
		{"id": "TPB", "name": "TPBank", "code": "970423", "short_name": "TPB", "logo": "https://img.vietqr.io/img/TPB.png", "color": "#6C2D8E"},
		{"id": "STB", "name": "Sacombank", "code": "970403", "short_name": "STB", "logo": "https://img.vietqr.io/img/STB.png", "color": "#0051A5"},
		{"id": "BIDV", "name": "BIDV", "code": "970418", "short_name": "BIDV", "logo": "https://img.vietqr.io/img/BIDV.png", "color": "#1B3A6B"},
		{"id": "VIB", "name": "VIB", "code": "970441", "short_name": "VIB", "logo": "https://img.vietqr.io/img/VIB.png", "color": "#1E3A8A"},
		{"id": "SHB", "name": "SHB", "code": "970443", "short_name": "SHB", "logo": "https://img.vietqr.io/img/SHB.png", "color": "#1D428A"},
		{"id": "CTG", "name": "VietinBank", "code": "970415", "short_name": "CTG", "logo": "https://img.vietqr.io/img/CTG.png", "color": "#004F9F"},
		{"id": "HDB", "name": "HDBank", "code": "970437", "short_name": "HDB", "logo": "https://img.vietqr.io/img/HDB.png", "color": "#E6332A"},
		{"id": "MSB", "name": "MSB", "code": "970426", "short_name": "MSB", "logo": "https://img.vietqr.io/img/MSB.png", "color": "#1A6DB0"},
		{"id": "EIB", "name": "Eximbank", "code": "970431", "short_name": "EIB", "logo": "https://img.vietqr.io/img/EIB.png", "color": "#0057A0"},
	}

	utils.RespondSuccess(c, http.StatusOK, "Supported banks", banks)
}

// generateVietQRURL creates a VietQR image URL
func generateVietQRURL(bankCode, accountNumber, accountName string, amount float64, description string) string {
	return fmt.Sprintf(
		"https://img.vietqr.io/image/%s-%s-compact2.png?amount=%.0f&addInfo=%s&accountName=%s",
		bankCode,
		accountNumber,
		amount,
		url.QueryEscape(description),
		url.QueryEscape(accountName),
	)
}

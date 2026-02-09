package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/repository"
	"github.com/splitbill/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionHandler struct {
	transactionRepo *repository.TransactionRepository
	userRepo        *repository.UserRepository
}

func NewTransactionHandler(transactionRepo *repository.TransactionRepository, userRepo *repository.UserRepository) *TransactionHandler {
	return &TransactionHandler{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}

// CreateTransaction godoc
// @Summary      Record a payment transaction
// @Description  Creates a new settlement transaction between two users in a group
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        request  body      models.CreateTransactionRequest  true  "Transaction details"
// @Success      201      {object}  utils.APIResponse{data=models.TransactionResponse}
// @Failure      400      {object}  utils.APIResponse
// @Failure      401      {object}  utils.APIResponse
// @Failure      500      {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /transactions [post]
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	// Get current user
	fromUser, err := h.userRepo.FindByFirebaseUID(c.Request.Context(), uid)
	if err != nil {
		utils.RespondUnauthorized(c, "User not found")
		return
	}

	toUserID, err := primitive.ObjectIDFromHex(req.ToUser)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid to_user ID")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(req.GroupID)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid group ID")
		return
	}

	tx := &models.Transaction{
		GroupID:         groupID,
		FromUser:        fromUser.ID,
		ToUser:          toUserID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Type:            models.TransactionSettlement,
		Status:          models.TransactionPending,
		PaymentMethod:   req.PaymentMethod,
		PaymentProofURL: req.PaymentProofURL,
		Note:            req.Note,
	}

	if req.BillID != "" {
		billID, err := primitive.ObjectIDFromHex(req.BillID)
		if err == nil {
			tx.BillID = billID
		}
	}

	if err := h.transactionRepo.Create(c.Request.Context(), tx); err != nil {
		utils.RespondInternalError(c, "Failed to create transaction: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, "Transaction recorded", tx.ToResponse())
}

// ConfirmTransaction godoc
// @Summary      Confirm a transaction
// @Description  Confirms a received payment. Only the recipient can confirm.
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {object}  utils.APIResponse
// @Failure      400  {object}  utils.APIResponse
// @Failure      401  {object}  utils.APIResponse
// @Failure      403  {object}  utils.APIResponse
// @Failure      404  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /transactions/{id}/confirm [put]
func (h *TransactionHandler) ConfirmTransaction(c *gin.Context) {
	txID := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(txID)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid transaction ID")
		return
	}

	// Verify the current user is the recipient
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	currentUser, err := h.userRepo.FindByFirebaseUID(c.Request.Context(), uid)
	if err != nil {
		utils.RespondUnauthorized(c, "User not found")
		return
	}

	tx, err := h.transactionRepo.FindByID(c.Request.Context(), objID)
	if err != nil {
		utils.RespondNotFound(c, "Transaction not found")
		return
	}

	if tx.ToUser != currentUser.ID {
		utils.RespondForbidden(c, "Only the recipient can confirm the transaction")
		return
	}

	if err := h.transactionRepo.Confirm(c.Request.Context(), objID); err != nil {
		utils.RespondInternalError(c, "Failed to confirm transaction")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Transaction confirmed", nil)
}

// GetUserDebts godoc
// @Summary      Get current user's debts
// @Description  Returns all transactions (debts) for the authenticated user across all groups
// @Tags         Users
// @Produce      json
// @Success      200  {object}  utils.APIResponse{data=[]models.TransactionResponse}
// @Failure      401  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /users/me/debts [get]
func (h *TransactionHandler) GetUserDebts(c *gin.Context) {
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	currentUser, err := h.userRepo.FindByFirebaseUID(c.Request.Context(), uid)
	if err != nil {
		utils.RespondUnauthorized(c, "User not found")
		return
	}

	transactions, err := h.transactionRepo.FindByUser(c.Request.Context(), currentUser.ID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to get debts")
		return
	}

	responses := make([]models.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		responses[i] = tx.ToResponse()
	}

	utils.RespondSuccess(c, http.StatusOK, "User debts", responses)
}

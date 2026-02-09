package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/services"
	"github.com/splitbill/backend/internal/utils"
)

type BillHandler struct {
	billService *services.BillService
	debtService *services.DebtService
}

func NewBillHandler(billService *services.BillService, debtService *services.DebtService) *BillHandler {
	return &BillHandler{
		billService: billService,
		debtService: debtService,
	}
}

// CreateBill godoc
// @Summary      Create a new bill
// @Description  Creates a new bill in a group with split information
// @Tags         Bills
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true  "Group ID"
// @Param        request  body      models.CreateBillRequest   true  "Bill creation data"
// @Success      201      {object}  utils.APIResponse{data=models.BillResponse}
// @Failure      400      {object}  utils.APIResponse
// @Failure      401      {object}  utils.APIResponse
// @Failure      500      {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id}/bills [post]
func (h *BillHandler) CreateBill(c *gin.Context) {
	var req models.CreateBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	groupID := c.Param("id")
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	bill, err := h.billService.CreateBill(c.Request.Context(), groupID, uid, req)
	if err != nil {
		utils.RespondInternalError(c, "Failed to create bill: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, "Bill created", bill.ToResponse())
}

// ListBills godoc
// @Summary      List group bills
// @Description  Returns all bills in a specific group
// @Tags         Bills
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  utils.APIResponse{data=[]models.BillResponse}
// @Failure      401  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id}/bills [get]
func (h *BillHandler) ListBills(c *gin.Context) {
	groupID := c.Param("id")

	bills, err := h.billService.ListBills(c.Request.Context(), groupID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to list bills: "+err.Error())
		return
	}

	responses := make([]models.BillResponse, len(bills))
	for i, b := range bills {
		responses[i] = b.ToResponse()
	}

	utils.RespondSuccess(c, http.StatusOK, "Bills retrieved", responses)
}

// GetBill godoc
// @Summary      Get bill details
// @Description  Returns detailed information about a specific bill
// @Tags         Bills
// @Produce      json
// @Param        id   path      string  true  "Bill ID"
// @Success      200  {object}  utils.APIResponse{data=models.BillResponse}
// @Failure      401  {object}  utils.APIResponse
// @Failure      404  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /bills/{id} [get]
func (h *BillHandler) GetBill(c *gin.Context) {
	billID := c.Param("id")

	bill, err := h.billService.GetBill(c.Request.Context(), billID)
	if err != nil {
		utils.RespondNotFound(c, "Bill not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Bill retrieved", bill.ToResponse())
}

// UpdateBill godoc
// @Summary      Update a bill
// @Description  Updates bill title, amount, splits, or category
// @Tags         Bills
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true  "Bill ID"
// @Param        request  body      models.UpdateBillRequest   true  "Bill update data"
// @Success      200      {object}  utils.APIResponse{data=models.BillResponse}
// @Failure      400      {object}  utils.APIResponse
// @Failure      401      {object}  utils.APIResponse
// @Failure      500      {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /bills/{id} [put]
func (h *BillHandler) UpdateBill(c *gin.Context) {
	var req models.UpdateBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	billID := c.Param("id")

	bill, err := h.billService.UpdateBill(c.Request.Context(), billID, req)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Bill updated", bill.ToResponse())
}

// DeleteBill godoc
// @Summary      Delete a bill
// @Description  Soft-deletes a bill (marks as deleted)
// @Tags         Bills
// @Produce      json
// @Param        id   path      string  true  "Bill ID"
// @Success      200  {object}  utils.APIResponse
// @Failure      401  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /bills/{id} [delete]
func (h *BillHandler) DeleteBill(c *gin.Context) {
	billID := c.Param("id")

	if err := h.billService.DeleteBill(c.Request.Context(), billID); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Bill deleted", nil)
}

// GetGroupBalances godoc
// @Summary      Get group balances
// @Description  Returns the balance for each member in the group (positive = owed, negative = owes)
// @Tags         Balances
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  utils.APIResponse{data=[]models.BalanceResponse}
// @Failure      401  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id}/balances [get]
func (h *BillHandler) GetGroupBalances(c *gin.Context) {
	groupID := c.Param("id")

	balances, err := h.debtService.GetGroupBalances(c.Request.Context(), groupID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to get balances: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Balances retrieved", balances)
}

// GetSettlements godoc
// @Summary      Get settlement suggestions
// @Description  Returns optimized settlement suggestions to minimize the number of transactions
// @Tags         Balances
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  utils.APIResponse{data=[]models.Settlement}
// @Failure      401  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id}/settlements [get]
func (h *BillHandler) GetSettlements(c *gin.Context) {
	groupID := c.Param("id")

	settlements, err := h.debtService.GetOptimalSettlements(c.Request.Context(), groupID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to get settlements: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Settlement suggestions", settlements)
}

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

// CreateBill creates a new bill in a group
// POST /api/v1/groups/:id/bills
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

// ListBills lists all bills in a group
// GET /api/v1/groups/:id/bills
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

// GetBill gets a specific bill
// GET /api/v1/bills/:id
func (h *BillHandler) GetBill(c *gin.Context) {
	billID := c.Param("id")

	bill, err := h.billService.GetBill(c.Request.Context(), billID)
	if err != nil {
		utils.RespondNotFound(c, "Bill not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Bill retrieved", bill.ToResponse())
}

// UpdateBill updates a bill
// PUT /api/v1/bills/:id
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

// DeleteBill deletes a bill
// DELETE /api/v1/bills/:id
func (h *BillHandler) DeleteBill(c *gin.Context) {
	billID := c.Param("id")

	if err := h.billService.DeleteBill(c.Request.Context(), billID); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Bill deleted", nil)
}

// GetGroupBalances gets all balances in a group
// GET /api/v1/groups/:id/balances
func (h *BillHandler) GetGroupBalances(c *gin.Context) {
	groupID := c.Param("id")

	balances, err := h.debtService.GetGroupBalances(c.Request.Context(), groupID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to get balances: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Balances retrieved", balances)
}

// GetSettlements gets optimal settlement suggestions
// GET /api/v1/groups/:id/settlements
func (h *BillHandler) GetSettlements(c *gin.Context) {
	groupID := c.Param("id")

	settlements, err := h.debtService.GetOptimalSettlements(c.Request.Context(), groupID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to get settlements: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Settlement suggestions", settlements)
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/splitbill/backend/internal/repository"
	"github.com/splitbill/backend/internal/services"
	"github.com/splitbill/backend/internal/utils"
)

type StatsHandler struct {
	statsService *services.StatsService
	userRepo     *repository.UserRepository
}

func NewStatsHandler(statsService *services.StatsService, userRepo *repository.UserRepository) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		userRepo:     userRepo,
	}
}

// GetGroupStats godoc
// @Summary      Get group statistics
// @Description  Returns comprehensive statistics for a group including spending, categories, and trends
// @Tags         Stats
// @Produce      json
// @Param        id  path      string  true  "Group ID"
// @Success      200 {object}  utils.APIResponse
// @Failure      400 {object}  utils.APIResponse
// @Failure      500 {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id}/stats [get]
func (h *StatsHandler) GetGroupStats(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		utils.RespondBadRequest(c, "Group ID is required")
		return
	}

	stats, err := h.statsService.GetGroupStats(c.Request.Context(), groupID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to get group stats: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Group stats retrieved", stats)
}

// GetUserStats godoc
// @Summary      Get current user's statistics
// @Description  Returns overall statistics for the authenticated user across all groups
// @Tags         Stats
// @Produce      json
// @Success      200 {object}  utils.APIResponse
// @Failure      401 {object}  utils.APIResponse
// @Failure      404 {object}  utils.APIResponse
// @Failure      500 {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /stats/me [get]
func (h *StatsHandler) GetUserStats(c *gin.Context) {
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	user, err := h.userRepo.FindByFirebaseUID(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		utils.RespondNotFound(c, "User not found")
		return
	}

	stats, err := h.statsService.GetUserStats(c.Request.Context(), user.ID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to get user stats: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "User stats retrieved", stats)
}

// ExportGroupSummary godoc
// @Summary      Export group summary
// @Description  Generates a text or JSON summary of a group's bills and settlements for sharing
// @Tags         Stats
// @Produce      json,plain
// @Param        id      path      string  true   "Group ID"
// @Param        format  query     string  false  "Export format: text or json (default text)"
// @Success      200     {object}  utils.APIResponse
// @Failure      400     {object}  utils.APIResponse
// @Failure      500     {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id}/export [get]
func (h *StatsHandler) ExportGroupSummary(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		utils.RespondBadRequest(c, "Group ID is required")
		return
	}

	format := c.DefaultQuery("format", "text")

	summary, err := h.statsService.ExportGroupSummary(c.Request.Context(), groupID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to export summary: "+err.Error())
		return
	}

	if format == "text" {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(summary))
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Export generated", gin.H{"summary": summary})
}

// GetCategoryList godoc
// @Summary      Get bill categories
// @Description  Returns the list of all supported bill categories with labels, icons, and colors
// @Tags         Categories
// @Produce      json
// @Success      200 {object}  utils.APIResponse
// @Router       /categories [get]
func (h *StatsHandler) GetCategoryList(c *gin.Context) {
	type CategoryInfo struct {
		Key   string `json:"key"`
		Label string `json:"label"`
		Icon  string `json:"icon"`
		Color string `json:"color"`
	}

	categories := []CategoryInfo{
		{Key: "food", Label: "Ăn uống", Icon: "restaurant", Color: "#FF6B6B"},
		{Key: "drinks", Label: "Đồ uống", Icon: "beer", Color: "#FFA502"},
		{Key: "groceries", Label: "Tạp hóa", Icon: "cart", Color: "#2ED573"},
		{Key: "transport", Label: "Di chuyển", Icon: "car", Color: "#1E90FF"},
		{Key: "accommodation", Label: "Chỗ ở", Icon: "bed", Color: "#A29BFE"},
		{Key: "entertainment", Label: "Giải trí", Icon: "game-controller", Color: "#FD79A8"},
		{Key: "shopping", Label: "Mua sắm", Icon: "bag-handle", Color: "#E17055"},
		{Key: "utilities", Label: "Tiện ích", Icon: "flash", Color: "#FDCB6E"},
		{Key: "health", Label: "Sức khỏe", Icon: "medkit", Color: "#00B894"},
		{Key: "travel", Label: "Du lịch", Icon: "airplane", Color: "#74B9FF"},
		{Key: "other", Label: "Khác", Icon: "ellipsis-horizontal", Color: "#636E72"},
	}

	utils.RespondSuccess(c, http.StatusOK, "Categories retrieved", categories)
}

// GetGroupCategoryStats godoc
// @Summary      Get group category statistics
// @Description  Returns category spending breakdown and monthly trends for a group
// @Tags         Stats
// @Produce      json
// @Param        id  path      string  true  "Group ID"
// @Success      200 {object}  utils.APIResponse
// @Failure      400 {object}  utils.APIResponse
// @Failure      500 {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id}/stats/categories [get]
func (h *StatsHandler) GetGroupCategoryStats(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		utils.RespondBadRequest(c, "Group ID is required")
		return
	}

	stats, err := h.statsService.GetGroupStats(c.Request.Context(), groupID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to get category stats: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Category stats retrieved", gin.H{
		"categories":    stats.CategoryStats,
		"monthly_trend": stats.MonthlyTrend,
		"total_spent":   stats.TotalSpent,
	})
}

package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/splitbill/backend/internal/repository"
	"github.com/splitbill/backend/internal/services"
	"github.com/splitbill/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityHandler struct {
	activityService *services.ActivityService
	userRepo        *repository.UserRepository
}

func NewActivityHandler(activityService *services.ActivityService, userRepo *repository.UserRepository) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
		userRepo:        userRepo,
	}
}

// GetGroupActivities returns activities for a specific group
// GET /api/v1/groups/:id/activities
func (h *ActivityHandler) GetGroupActivities(c *gin.Context) {
	groupID := c.Param("id")
	if _, err := primitive.ObjectIDFromHex(groupID); err != nil {
		utils.RespondBadRequest(c, "Invalid group ID")
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 {
		limit = 20
	}

	activities, err := h.activityService.GetGroupActivities(c.Request.Context(), groupID, limit)
	if err != nil {
		utils.RespondInternalError(c, "Failed to get activities: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Group activities", activities)
}

// GetUserActivities returns activities across all user's groups
// GET /api/v1/activities/me
func (h *ActivityHandler) GetUserActivities(c *gin.Context) {
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	user, err := h.userRepo.FindByFirebaseUID(c.Request.Context(), uid)
	if err != nil {
		utils.RespondUnauthorized(c, "User not found")
		return
	}

	limitStr := c.DefaultQuery("limit", "30")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 {
		limit = 30
	}

	activities, err := h.activityService.GetUserActivities(c.Request.Context(), user.ID, limit)
	if err != nil {
		utils.RespondInternalError(c, "Failed to get activities: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "User activities", activities)
}

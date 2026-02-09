package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/services"
	"github.com/splitbill/backend/internal/utils"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// VerifyToken verifies Firebase token and returns/creates user
// POST /api/v1/auth/verify-token
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	firebaseUID, _ := c.Get("firebase_uid")
	phone, _ := c.Get("user_phone")

	uid := firebaseUID.(string)
	phoneStr := ""
	if phone != nil {
		phoneStr = phone.(string)
	}

	user, err := h.authService.VerifyAndGetUser(c.Request.Context(), uid, phoneStr)
	if err != nil {
		utils.RespondInternalError(c, "Failed to verify user: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "User verified", user.ToResponse())
}

// GetMe returns the current user's profile
// GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	user, err := h.authService.GetUserByFirebaseUID(c.Request.Context(), uid)
	if err != nil {
		utils.RespondNotFound(c, "User not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "User profile", user.ToResponse())
}

// UpdateProfile updates the current user's profile
// PUT /api/v1/auth/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	user, err := h.authService.UpdateProfile(c.Request.Context(), uid, req)
	if err != nil {
		utils.RespondInternalError(c, "Failed to update profile: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Profile updated", user.ToResponse())
}

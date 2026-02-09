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

// VerifyToken godoc
// @Summary      Verify Firebase token
// @Description  Verifies Firebase ID token and returns or creates the user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.APIResponse{data=models.UserResponse}
// @Failure      401  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /auth/verify-token [post]
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

// GetMe godoc
// @Summary      Get current user profile
// @Description  Returns the authenticated user's profile information
// @Tags         Auth
// @Produce      json
// @Success      200  {object}  utils.APIResponse{data=models.UserResponse}
// @Failure      401  {object}  utils.APIResponse
// @Failure      404  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /auth/me [get]
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

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates the authenticated user's profile (display name, avatar, bank info)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.UpdateUserRequest  true  "Profile update data"
// @Success      200      {object}  utils.APIResponse{data=models.UserResponse}
// @Failure      400      {object}  utils.APIResponse
// @Failure      401      {object}  utils.APIResponse
// @Failure      500      {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /auth/profile [put]
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

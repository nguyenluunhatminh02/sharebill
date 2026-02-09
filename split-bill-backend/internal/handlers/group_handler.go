package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/services"
	"github.com/splitbill/backend/internal/utils"
)

type GroupHandler struct {
	groupService *services.GroupService
}

func NewGroupHandler(groupService *services.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

// CreateGroup godoc
// @Summary      Create a new group
// @Description  Creates a new group with the authenticated user as creator and first member
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Param        request  body      models.CreateGroupRequest  true  "Group creation data"
// @Success      201      {object}  utils.APIResponse{data=models.GroupResponse}
// @Failure      400      {object}  utils.APIResponse
// @Failure      401      {object}  utils.APIResponse
// @Failure      500      {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req models.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	group, err := h.groupService.CreateGroup(c.Request.Context(), uid, req)
	if err != nil {
		utils.RespondInternalError(c, "Failed to create group: "+err.Error())
		return
	}

	resp, _ := h.groupService.GetGroupWithMemberDetails(c.Request.Context(), group)
	utils.RespondSuccess(c, http.StatusCreated, "Group created", resp)
}

// ListGroups godoc
// @Summary      List user's groups
// @Description  Returns all groups the authenticated user is a member of
// @Tags         Groups
// @Produce      json
// @Success      200  {object}  utils.APIResponse{data=[]models.GroupResponse}
// @Failure      401  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups [get]
func (h *GroupHandler) ListGroups(c *gin.Context) {
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	groups, err := h.groupService.ListGroups(c.Request.Context(), uid)
	if err != nil {
		utils.RespondInternalError(c, "Failed to list groups: "+err.Error())
		return
	}

	responses := make([]models.GroupResponse, len(groups))
	for i, g := range groups {
		responses[i] = g.ToResponse()
	}

	utils.RespondSuccess(c, http.StatusOK, "Groups retrieved", responses)
}

// GetGroup godoc
// @Summary      Get group details
// @Description  Returns detailed information about a specific group including member details
// @Tags         Groups
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  utils.APIResponse{data=models.GroupResponse}
// @Failure      401  {object}  utils.APIResponse
// @Failure      404  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id} [get]
func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupID := c.Param("id")
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	group, err := h.groupService.GetGroup(c.Request.Context(), groupID, uid)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	resp, _ := h.groupService.GetGroupWithMemberDetails(c.Request.Context(), group)
	utils.RespondSuccess(c, http.StatusOK, "Group retrieved", resp)
}

// UpdateGroup godoc
// @Summary      Update a group
// @Description  Updates group name, description, or currency. Only the group creator can update.
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true  "Group ID"
// @Param        request  body      models.UpdateGroupRequest  true  "Group update data"
// @Success      200      {object}  utils.APIResponse{data=models.GroupResponse}
// @Failure      400      {object}  utils.APIResponse
// @Failure      401      {object}  utils.APIResponse
// @Failure      500      {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id} [put]
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	var req models.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	groupID := c.Param("id")
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	group, err := h.groupService.UpdateGroup(c.Request.Context(), groupID, uid, req)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Group updated", group.ToResponse())
}

// DeleteGroup godoc
// @Summary      Delete a group
// @Description  Deletes a group. Only the group creator can delete.
// @Tags         Groups
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  utils.APIResponse
// @Failure      401  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id} [delete]
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	if err := h.groupService.DeleteGroup(c.Request.Context(), groupID, uid); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Group deleted", nil)
}

// AddMember godoc
// @Summary      Add member to group
// @Description  Adds a user to a group by user ID
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Param        id       path      string                    true  "Group ID"
// @Param        request  body      models.AddMemberRequest   true  "Member data"
// @Success      200      {object}  utils.APIResponse
// @Failure      400      {object}  utils.APIResponse
// @Failure      401      {object}  utils.APIResponse
// @Failure      500      {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id}/members [post]
func (h *GroupHandler) AddMember(c *gin.Context) {
	var req models.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	groupID := c.Param("id")
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	if err := h.groupService.AddMember(c.Request.Context(), groupID, uid, req); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Member added", nil)
}

// RemoveMember godoc
// @Summary      Remove member from group
// @Description  Removes a user from a group. Only creator or the member themselves can remove.
// @Tags         Groups
// @Produce      json
// @Param        id      path      string  true  "Group ID"
// @Param        userId  path      string  true  "User ID to remove"
// @Success      200     {object}  utils.APIResponse
// @Failure      401     {object}  utils.APIResponse
// @Failure      500     {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/{id}/members/{userId} [delete]
func (h *GroupHandler) RemoveMember(c *gin.Context) {
	groupID := c.Param("id")
	memberUserID := c.Param("userId")
	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	if err := h.groupService.RemoveMember(c.Request.Context(), groupID, uid, memberUserID); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Member removed", nil)
}

// JoinGroup godoc
// @Summary      Join group via invite code
// @Description  Joins a group using the group's invite code
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Param        request  body      models.JoinGroupRequest  true  "Invite code"
// @Success      200      {object}  utils.APIResponse{data=models.GroupResponse}
// @Failure      400      {object}  utils.APIResponse
// @Failure      401      {object}  utils.APIResponse
// @Failure      500      {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /groups/join [post]
func (h *GroupHandler) JoinGroup(c *gin.Context) {
	var req models.JoinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	firebaseUID, _ := c.Get("firebase_uid")
	uid := firebaseUID.(string)

	group, err := h.groupService.JoinGroup(c.Request.Context(), uid, req.InviteCode)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	resp, _ := h.groupService.GetGroupWithMemberDetails(c.Request.Context(), group)
	utils.RespondSuccess(c, http.StatusOK, "Joined group", resp)
}

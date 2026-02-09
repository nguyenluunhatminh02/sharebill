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

// CreateGroup creates a new group
// POST /api/v1/groups
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

// ListGroups lists all groups for the current user
// GET /api/v1/groups
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

// GetGroup gets a specific group
// GET /api/v1/groups/:id
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

// UpdateGroup updates a group
// PUT /api/v1/groups/:id
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

// DeleteGroup deletes a group
// DELETE /api/v1/groups/:id
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

// AddMember adds a member to a group
// POST /api/v1/groups/:id/members
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

// RemoveMember removes a member from a group
// DELETE /api/v1/groups/:id/members/:userId
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

// JoinGroup joins a group using an invite code
// POST /api/v1/groups/join
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

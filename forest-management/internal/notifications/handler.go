package notifications

import (
	"strconv"

	"forest-management/internal/audit"
	"forest-management/pkg/middleware"
	"forest-management/pkg/requestutil"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service *NotificationService
}

func NewNotificationHandler(service *NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// List — Get user's notifications
func (h *NotificationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	page, perPage := requestutil.Pagination(c)

	notifications, meta, err := h.service.GetUserNotifications(userID, role, page, perPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch notifications")
		return
	}

	response.Paginated(c, "Notifications retrieved", notifications, (*response.Pagination)(meta))
}

// UnreadCount — Get unread notification count
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	count, err := h.service.GetUnreadCount(userID, role)
	if err != nil {
		response.InternalError(c, "Failed to get unread count")
		return
	}

	response.Success(c, "Unread count", gin.H{"unread_count": count})
}

// MarkRead — Mark a notification as read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, _ := strconv.Atoi(c.Param("id"))

	role := middleware.GetUserRole(c)

	if err := h.service.MarkAsRead(uint(id), userID, role); err != nil {
		response.Error(c, 500, "Failed to mark as read")
		return
	}

	response.Success(c, "Notification marked as read", nil)
}

// MarkAllRead — Mark all notifications as read
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	if err := h.service.MarkAllAsRead(userID, role); err != nil {
		response.Error(c, 500, "Failed to mark all as read")
		return
	}

	response.Success(c, "All notifications marked as read", nil)
}

// AdminCreate — Admin creates a notification (broadcast or targeted)
func (h *NotificationHandler) AdminCreate(c *gin.Context) {
	var input struct {
		UserID     *uint   `json:"user_id"`
		TargetRole *string `json:"target_role"` // admin, staff, member
		Title      string  `json:"title" binding:"required"`
		Message    string  `json:"message" binding:"required"`
		Type       string  `json:"type" binding:"required"` // info, warning, success, system
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}

	notif, err := h.service.CreateNotification(input.UserID, input.TargetRole, input.Title, input.Message, input.Type, nil, nil)
	if err != nil {
		response.Error(c, 500, "Failed to create notification")
		return
	}

	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(
		h.service.db,
		&actorID,
		"create",
		"notification",
		&notif.ID,
		nil,
		gin.H{
			"user_id":     input.UserID,
			"target_role": input.TargetRole,
			"title":       input.Title,
			"type":        input.Type,
		},
		c.ClientIP(),
		c.Request.UserAgent(),
		"Administrator created a notification",
	)

	response.Created(c, "Notification created", notif)
}

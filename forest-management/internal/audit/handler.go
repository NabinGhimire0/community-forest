package audit

import (
	"strconv"

	"forest-management/pkg/requestutil"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service *AuditService
}

func NewAuditHandler(service *AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

// List — Admin views audit logs
func (h *AuditHandler) List(c *gin.Context) {
	page, perPage := requestutil.Pagination(c)
	action := c.Query("action")
	entity := c.Query("entity")
	userID := c.Query("user_id")

	logs, meta, err := h.service.ListLogs(page, perPage, action, entity, userID)
	if err != nil {
		response.InternalError(c, "Failed to fetch audit logs")
		return
	}

	response.Paginated(c, "Audit logs retrieved", logs, (*response.Pagination)(meta))
}

// EntityHistory — View change history for a specific entity
func (h *AuditHandler) EntityHistory(c *gin.Context) {
	entity := c.Query("entity")
	entityID, _ := strconv.Atoi(c.Query("entity_id"))

	if entity == "" || entityID == 0 {
		response.BadRequest(c, "entity and entity_id are required")
		return
	}

	logs, err := h.service.GetEntityHistory(entity, uint(entityID))
	if err != nil {
		response.InternalError(c, "Failed to fetch entity history")
		return
	}

	response.Success(c, "Entity history retrieved", logs)
}

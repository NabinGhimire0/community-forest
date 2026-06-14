package members

import (
	"forest-management/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMemberFinancialSummary returns financial summary for a member
func (h *MemberHandler) GetMemberFinancialSummary(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	summary, err := h.service.GetMemberFinancialSummary(uint(memberID))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Financial summary retrieved", summary)
}

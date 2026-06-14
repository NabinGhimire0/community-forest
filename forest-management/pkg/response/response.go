package response

import (
	"log"
	"net/http"

	"forest-management/config"

	"github.com/gin-gonic/gin"
)

// Response is the standard JSON envelope for all API responses.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse adds pagination metadata.
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    *Pagination `json:"meta,omitempty"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Message: message, Data: data})
}

// Error sends an error response. Unexpected server details are never exposed in
// production; the full diagnostic is written to server logs together with the
// request identifier so operators can investigate without leaking internals.
func Error(c *gin.Context, statusCode int, message string) {
	publicMessage := message
	publicError := message
	if statusCode >= http.StatusInternalServerError {
		requestID, _ := c.Get("request_id")
		log.Printf("request_id=%v status=%d internal_error=%q", requestID, statusCode, message)
		if config.AppConfig != nil && config.AppConfig.AppEnv == "production" {
			publicMessage = "An unexpected server error occurred"
			publicError = "internal_server_error"
		}
	}
	c.JSON(statusCode, Response{Success: false, Message: publicMessage, Error: publicError})
}

func BadRequest(c *gin.Context, message string)    { Error(c, http.StatusBadRequest, message) }
func Unauthorized(c *gin.Context, message string)  { Error(c, http.StatusUnauthorized, message) }
func Forbidden(c *gin.Context, message string)     { Error(c, http.StatusForbidden, message) }
func NotFound(c *gin.Context, message string)      { Error(c, http.StatusNotFound, message) }
func InternalError(c *gin.Context, message string) { Error(c, http.StatusInternalServerError, message) }

func Paginated(c *gin.Context, message string, data interface{}, meta *Pagination) {
	c.JSON(http.StatusOK, PaginatedResponse{Success: true, Message: message, Data: data, Meta: meta})
}

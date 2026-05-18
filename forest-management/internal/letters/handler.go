package letters

import (
	"strconv"

	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type LetterHandler struct {
	service *LetterService
}

func NewLetterHandler(service *LetterService) *LetterHandler {
	return &LetterHandler{service: service}
}

type CreateLetterInput struct {
	Type         string  `json:"type" binding:"required,oneof=incoming outgoing"`
	ReferenceNo  *string `json:"reference_no"`
	Title        string  `json:"title" binding:"required"`
	Subject      string  `json:"subject" binding:"required"`
	FromParty    *string `json:"from_party"`
	ToParty      *string `json:"to_party"`
	LetterDate   string  `json:"letter_date" binding:"required"`
	ReceivedDate *string `json:"received_date"`
	SentDate     *string `json:"sent_date"`
	DocumentFile *string `json:"document_file"`
	Remarks      *string `json:"remarks"`
}

type UpdateLetterInput struct {
	Type         *string `json:"type"`
	ReferenceNo  *string `json:"reference_no"`
	Title        *string `json:"title"`
	Subject      *string `json:"subject"`
	FromParty    *string `json:"from_party"`
	ToParty      *string `json:"to_party"`
	LetterDate   *string `json:"letter_date"`
	ReceivedDate *string `json:"received_date"`
	SentDate     *string `json:"sent_date"`
	DocumentFile *string `json:"document_file"`
	Remarks      *string `json:"remarks"`
}

func (h *LetterHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var input CreateLetterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid letter data: "+err.Error())
		return
	}

	letter, err := h.service.CreateLetter(userID, input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Created(c, "Letter recorded successfully", letter)
}

func (h *LetterHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	letterType := c.Query("type")
	search := c.Query("search")

	letters, meta, err := h.service.ListLetters(page, perPage, letterType, search)
	if err != nil {
		response.InternalError(c, "Failed to fetch letters")
		return
	}

	response.Paginated(c, "Letters retrieved", letters, meta)
}

func (h *LetterHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid letter ID")
		return
	}

	letter, err := h.service.GetLetterByID(uint(id))
	if err != nil {
		response.NotFound(c, "Letter not found")
		return
	}

	response.Success(c, "Letter retrieved", letter)
}

func (h *LetterHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid letter ID")
		return
	}
	userID := middleware.GetUserID(c)

	var input UpdateLetterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid letter data")
		return
	}

	letter, err := h.service.UpdateLetter(uint(id), userID, input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Letter updated", letter)
}

func (h *LetterHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid letter ID")
		return
	}

	if err := h.service.DeleteLetter(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Letter deleted", nil)
}

// UploadDocument handles file upload for letter documents
func (h *LetterHandler) UploadDocument(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid letter ID")
		return
	}

	file, header, err := c.Request.FormFile("document")
	if err != nil {
		response.BadRequest(c, "Document file is required")
		return
	}
	defer file.Close()

	// Validate file type
	allowedTypes := map[string]bool{
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"image/jpeg": true,
		"image/png":  true,
	}
	mimeType := header.Header.Get("Content-Type")
	if !allowedTypes[mimeType] {
		response.BadRequest(c, "Only PDF, DOC, DOCX, JPEG, PNG files are allowed")
		return
	}

	// Max 10MB
	if header.Size > 10*1024*1024 {
		response.BadRequest(c, "File size must be less than 10MB")
		return
	}

	documentURL, err := h.service.UploadDocument(uint(id), file, header.Filename)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Document uploaded successfully", gin.H{
		"document_url": documentURL,
	})
}

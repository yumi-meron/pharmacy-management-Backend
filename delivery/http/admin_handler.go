package http

import (
	"net/http"

	"pharmacist-backend/domain"
	"pharmacist-backend/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	usecase   usecase.AdminUsecase
	validator *validator.Validate
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(usecase usecase.AdminUsecase, validator *validator.Validate) *AdminHandler {
	return &AdminHandler{usecase, validator}
}

// CreateUser handles POST /admin/users to create a user
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var input domain.SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user
	if err := h.usecase.CreateOwner(c.Request.Context(), input); err != nil {
		switch err {
		case domain.ErrPhoneNumberTaken:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case domain.ErrInvalidRole:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

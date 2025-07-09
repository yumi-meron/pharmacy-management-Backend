package http

import (
	"net/http"

	"pharmacist-backend/domain"
	"pharmacist-backend/usecase"
	"pharmacist-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	usecase   usecase.UserUsecase
	validator *validator.Validate
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(usecase usecase.UserUsecase, validator *validator.Validate) *UserHandler {
	return &UserHandler{usecase, validator}
}

// CreateOwner handles POST /api/users/owners
func (h *UserHandler) CreateOwner(c *gin.Context) {
	var input domain.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	role, _ := c.Get("role")
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	if err := h.usecase.CreateOwner(c.Request.Context(), role.(string), userID, input); err != nil {
		switch err {
		case domain.ErrPhoneNumberTaken:
			utils.ErrorResponse(c, http.StatusConflict, err)
		case domain.ErrUnauthorized, domain.ErrInvalidRole:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Owner created successfully"})
}

// CreatePharmacist handles POST /api/users/pharmacists
func (h *UserHandler) CreatePharmacist(c *gin.Context) {
	var input domain.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	role, _ := c.Get("role")
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	if err := h.usecase.CreatePharmacist(c.Request.Context(), role.(string), userID, input); err != nil {
		switch err {
		case domain.ErrPhoneNumberTaken:
			utils.ErrorResponse(c, http.StatusConflict, err)
		case domain.ErrUnauthorized, domain.ErrInvalidRole:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pharmacist created successfully"})
}

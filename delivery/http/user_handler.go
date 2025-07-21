package http

import (
	"errors"
	"net/http"

	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/usecase"
	"pharmacy-management-backend/utils"

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

// ListPharmacists handles GET /api/users/pharmacists
func (h *UserHandler) ListPharmacists(c *gin.Context) {
	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, err := uuid.Parse(pharmacyIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid pharmacy ID"))
		return
	}

	pharmacists, err := h.usecase.ListPharmacists(c.Request.Context(), role.(string), pharmacyID)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	// Format response to exclude password
	response := make([]gin.H, len(pharmacists))
	for i, user := range pharmacists {
		response[i] = gin.H{
			"id":              user.ID,
			"phone_number":    user.PhoneNumber,
			"full_name":       user.FullName,
			"role":            user.Role,
			"pharmacy_id":     user.PharmacyID,
			"profile_picture": user.ProfilePicture,
			"created_at":      user.CreatedAt,
			"updated_at":      user.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

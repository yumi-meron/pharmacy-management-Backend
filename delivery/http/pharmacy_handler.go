package http

import (
	"errors"
	"net/http"

	"pharmacist-backend/domain"
	"pharmacist-backend/usecase"
	"pharmacist-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// PharmacyHandler handles pharmacy-related HTTP requests
type PharmacyHandler struct {
	usecase   usecase.PharmacyUsecase
	validator *validator.Validate
}

// NewPharmacyHandler creates a new PharmacyHandler
func NewPharmacyHandler(usecase usecase.PharmacyUsecase, validator *validator.Validate) *PharmacyHandler {
	return &PharmacyHandler{usecase, validator}
}

// Create handles POST /api/pharmacies
func (h *PharmacyHandler) Create(c *gin.Context) {
	var input domain.Pharmacy
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
	if err := h.usecase.Create(c.Request.Context(), role.(string), input); err != nil {
		switch err {
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pharmacy created successfully"})
}

// GetAll handles GET /api/pharmacies
func (h *PharmacyHandler) GetAll(c *gin.Context) {
	role, _ := c.Get("role")
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	pharmacies, err := h.usecase.GetAll(c.Request.Context(), role.(string), userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, pharmacies)
}

// GetByID handles GET /api/pharmacies/:id
func (h *PharmacyHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid pharmacy ID"))
		return
	}

	role, _ := c.Get("role")
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	pharmacy, err := h.usecase.GetByID(c.Request.Context(), role.(string), userID, id)
	if err != nil {
		switch err {
		case domain.ErrNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, pharmacy)
}

// Update handles PUT /api/pharmacies/:id
func (h *PharmacyHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid pharmacy ID"))
		return
	}

	var input domain.Pharmacy
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

	if err := h.usecase.Update(c.Request.Context(), role.(string), userID, id, input); err != nil {
		switch err {
		case domain.ErrNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pharmacy updated successfully"})
}

// Delete handles DELETE /api/pharmacies/:id
func (h *PharmacyHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid pharmacy ID"))
		return
	}

	role, _ := c.Get("role")
	if err := h.usecase.Delete(c.Request.Context(), role.(string), id); err != nil {
		switch err {
		case domain.ErrNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pharmacy deleted successfully"})
}

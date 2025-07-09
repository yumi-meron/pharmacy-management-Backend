package http

import (
	"net/http"
	"strconv"

	"pharmacist-backend/domain"
	"pharmacist-backend/usecase"

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

// Create handles POST /pharmacies
func (h *PharmacyHandler) Create(c *gin.Context) {
	var input domain.Pharmacy
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create pharmacy
	if err := h.usecase.Create(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create pharmacy"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pharmacy created successfully"})
}

// GetByID handles GET /pharmacies/:id
func (h *PharmacyHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pharmacy ID"})
		return
	}

	// Retrieve pharmacy
	pharmacy, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrPharmacyNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve pharmacy"})
		}
		return
	}

	c.JSON(http.StatusOK, pharmacy)
}

// GetAll handles GET /pharmacies
func (h *PharmacyHandler) GetAll(c *gin.Context) {
	// Parse query parameters for pagination
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Retrieve pharmacies
	pharmacies, err := h.usecase.GetAll(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve pharmacies"})
		return
	}

	c.JSON(http.StatusOK, pharmacies)
}

// Update handles PUT /pharmacies/:id
func (h *PharmacyHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pharmacy ID"})
		return
	}

	var input domain.Pharmacy
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set ID from URL
	input.ID = id

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update pharmacy
	if err := h.usecase.Update(c.Request.Context(), input); err != nil {
		switch err {
		case domain.ErrPharmacyNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update pharmacy"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pharmacy updated successfully"})
}

// Delete handles DELETE /pharmacies/:id
func (h *PharmacyHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pharmacy ID"})
		return
	}

	// Delete pharmacy
	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		switch err {
		case domain.ErrPharmacyNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete pharmacy"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pharmacy deleted successfully"})
}

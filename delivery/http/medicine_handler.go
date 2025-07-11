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

// MedicineHandler handles medicine-related HTTP requests
type MedicineHandler struct {
	usecase   usecase.MedicineUsecase
	validator *validator.Validate
}

// NewMedicineHandler creates a new MedicineHandler
func NewMedicineHandler(usecase usecase.MedicineUsecase, validator *validator.Validate) *MedicineHandler {
	return &MedicineHandler{usecase, validator}
}

// Create handles POST /api/medicines
func (h *MedicineHandler) Create(c *gin.Context) {
	var input domain.CreateMedicineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	if err := h.usecase.Create(c.Request.Context(), role.(string), pharmacyID, input); err != nil {
		switch err {
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		case domain.ErrInvalidPharmacy:
			utils.ErrorResponse(c, http.StatusBadRequest, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Medicine created successfully"})
}

// GetAll handles GET /api/medicines
func (h *MedicineHandler) GetAll(c *gin.Context) {
	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	medicines, err := h.usecase.GetAll(c.Request.Context(), role.(string), pharmacyID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, medicines)
}

// GetByID handles GET /api/medicines/:id
func (h *MedicineHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid medicine ID"))
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	medicine, err := h.usecase.GetByID(c.Request.Context(), role.(string), pharmacyID, id)
	if err != nil {
		switch err {
		case domain.ErrMedicineNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, medicine)
}

// Update handles PUT /api/medicines/:id
func (h *MedicineHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid medicine ID"))
		return
	}

	var input domain.UpdateMedicineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	if err := h.usecase.Update(c.Request.Context(), role.(string), pharmacyID, id, input); err != nil {
		switch err {
		case domain.ErrMedicineNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Medicine updated successfully"})
}

// Delete handles DELETE /api/medicines/:id
func (h *MedicineHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid medicine ID"))
		return
	}

	role, _ := c.Get("role")
	if err := h.usecase.Delete(c.Request.Context(), role.(string), id); err != nil {
		switch err {
		case domain.ErrMedicineNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		case domain.ErrMedicineHasVariants:
			utils.ErrorResponse(c, http.StatusBadRequest, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Medicine deleted successfully"})
}

// CreateVariant handles POST /api/medicines/:id/variants
func (h *MedicineHandler) CreateVariant(c *gin.Context) {
	idStr := c.Param("id")
	medicineID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid medicine ID"))
		return
	}

	var input domain.CreateMedicineVariantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	if err := h.usecase.CreateVariant(c.Request.Context(), role.(string), pharmacyID, medicineID, input); err != nil {
		switch err {
		case domain.ErrMedicineNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		case domain.ErrBarcodeTaken:
			utils.ErrorResponse(c, http.StatusConflict, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Medicine variant created successfully"})
}

// GetVariants handles GET /api/medicines/:id/variants
func (h *MedicineHandler) GetVariants(c *gin.Context) {
	idStr := c.Param("id")
	medicineID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid medicine ID"))
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	variants, err := h.usecase.GetVariants(c.Request.Context(), role.(string), pharmacyID, medicineID)
	if err != nil {
		switch err {
		case domain.ErrMedicineNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, variants)
}

// GetVariantByID handles GET /api/medicines/:id/variants/:variant_id
func (h *MedicineHandler) GetVariantByID(c *gin.Context) {
	idStr := c.Param("id")
	medicineID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid medicine ID"))
		return
	}

	variantIDStr := c.Param("variant_id")
	variantID, err := uuid.Parse(variantIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid variant ID"))
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	variant, err := h.usecase.GetVariantByID(c.Request.Context(), role.(string), pharmacyID, medicineID, variantID)
	if err != nil {
		switch err {
		case domain.ErrVariantNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, variant)
}

// UpdateVariant handles PUT /api/medicines/:id/variants/:variant_id
func (h *MedicineHandler) UpdateVariant(c *gin.Context) {
	idStr := c.Param("id")
	medicineID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid medicine ID"))
		return
	}

	variantIDStr := c.Param("variant_id")
	variantID, err := uuid.Parse(variantIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid variant ID"))
		return
	}

	var input domain.UpdateMedicineVariantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	if err := h.usecase.UpdateVariant(c.Request.Context(), role.(string), pharmacyID, medicineID, variantID, input); err != nil {
		switch err {
		case domain.ErrVariantNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		case domain.ErrBarcodeTaken:
			utils.ErrorResponse(c, http.StatusConflict, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Medicine variant updated successfully"})
}

// DeleteVariant handles DELETE /api/medicines/:id/variants/:variant_id
func (h *MedicineHandler) DeleteVariant(c *gin.Context) {
	variantIDStr := c.Param("variant_id")
	variantID, err := uuid.Parse(variantIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid variant ID"))
		return
	}

	role, _ := c.Get("role")
	if err := h.usecase.DeleteVariant(c.Request.Context(), role.(string), variantID); err != nil {
		switch err {
		case domain.ErrVariantNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Medicine variant deleted successfully"})
}

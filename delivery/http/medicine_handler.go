package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/usecase"
	"pharmacy-management-backend/utils"

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
// func (h *MedicineHandler) Update(c *gin.Context) {
// 	idStr := c.Param("id")
// 	id, err := uuid.Parse(idStr)
// 	if err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid medicine ID"))
// 		return
// 	}

// 	var input domain.UpdateMedicineInput
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	if err := h.validator.Struct(input); err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	role, _ := c.Get("role")
// 	pharmacyIDStr, _ := c.Get("pharmacy_id")
// 	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

// 	if err := h.usecase.Update(c.Request.Context(), role.(string), pharmacyID, id, input); err != nil {
// 		switch err {
// 		case domain.ErrMedicineNotFound:
// 			utils.ErrorResponse(c, http.StatusNotFound, err)
// 		case domain.ErrUnauthorized:
// 			utils.ErrorResponse(c, http.StatusForbidden, err)
// 		default:
// 			utils.ErrorResponse(c, http.StatusInternalServerError, err)
// 		}
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Medicine updated successfully"})
// }

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
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Errorf("error parsing form: %v", err))
		return
	}

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

	var input = domain.UpdateMedicineVariantInput{
		Name:         c.PostForm("name"),
		Brand:        c.PostForm("brand"),
		Description:  c.PostForm("description"),
		Barcode:      c.PostForm("barcode"),
		Unit:         c.PostForm("unit"),
		PricePerUnit: parseFloat(c.PostForm("price_per_unit")),
		Stock:        parseInt(c.PostForm("stock")),
		ExpiryDate:   parseDate(c.PostForm("expiry_date")),
		Picture:      "",
	}

	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Errorf("validation failed: %v", err))
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	file, fileHeader, err := c.Request.FormFile("picture")
	var pictureURL string
	if err == nil && file != nil && fileHeader != nil {
		defer file.Close()
		fileData, err := io.ReadAll(file)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("error reading file: %v", err))
			return
		}
		if len(fileData) == 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, fmt.Errorf("empty file uploaded"))
			return
		}

		// Extract file extension
		fileExt := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if fileExt == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, fmt.Errorf("file has no extension"))
			return
		}

		pictureURL, err = h.usecase.UploadPicture(c.Request.Context(), variantID, fileData, fileExt)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to upload variant picture: %v", err))
			return
		}
		input.Picture = pictureURL
	}

	medicine, err := h.usecase.UpdateVariant(c.Request.Context(), role.(string), pharmacyID, medicineID, variantID, input)
	if err != nil {
		switch err {
		case domain.ErrVariantNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		case domain.ErrBarcodeTaken:
			utils.ErrorResponse(c, http.StatusConflict, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to update variant: %v", err))
		}
		return
	}

	c.JSON(http.StatusOK, medicine)
}

// Helper functions for parsing form data
func parseFloat(value string) float64 {
	if value == "" {
		return 0
	}
	result, _ := strconv.ParseFloat(value, 64)
	return result
}

func parseInt(value string) int {
	if value == "" {
		return 0
	}
	result, _ := strconv.Atoi(value)
	return result
}

func parseDate(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsedDate, _ := time.Parse("2006-01-02", value)
	return parsedDate
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

func (h *MedicineHandler) SearchMedicines(c *gin.Context) {
	var params domain.SearchParams

	// Try binding JSON body first
	if err := c.ShouldBindJSON(&params); err != nil {
		// If JSON binding fails, try query parameters as a fallback
		params.Query = c.Query("query")
		params.Filter = c.Query("filter")

		if limitStr := c.Query("limit"); limitStr != "" {
			var limit int
			if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
				return
			}
			params.Limit = limit
		}
		if offsetStr := c.Query("offset"); offsetStr != "" {
			var offset int
			if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
				return
			}
			params.Offset = offset
		}
	}

	// Parse PharmacyID from context (set by middleware)
	pharmacyIDStr, exists := c.Get("pharmacy_id")
	if exists {
		pharmacyID, err := uuid.Parse(pharmacyIDStr.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pharmacy ID"})
			return
		}
		params.PharmacyID = pharmacyID
	}

	// Validate SearchParams
	if err := h.validator.Struct(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("validation failed: %v", err)})
		return
	}

	// Execute search
	medicines, err := h.usecase.SearchMedicines(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to search medicines: %v", err)})
		return
	}

	// Respond with JSON
	c.JSON(http.StatusOK, medicines)
}

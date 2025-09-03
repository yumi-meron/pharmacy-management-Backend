package http

import (
	"errors"
	"net/http"

	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/usecase"
	"pharmacy-management-backend/utils"

	"github.com/gin-gonic/gin"
)

// BarcodeHandler handles barcode-related HTTP requests
type BarcodeHandler struct {
	usecase usecase.BarcodeUsecase
}

// NewBarcodeHandler creates a new BarcodeHandler
func NewBarcodeHandler(usecase usecase.BarcodeUsecase) *BarcodeHandler {
	return &BarcodeHandler{usecase}
}

// LookupBarcode handles GET /api/barcodes?value={barcodeValue}
func (h *BarcodeHandler) LookupBarcode(c *gin.Context) {
	// Get barcode value from the query parameter
	barcodeValue := c.Param("barcode_value")
	if barcodeValue == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("barcode value is required"))
		return
	}

	// Lookup barcode using the usecase
	medicine, err := h.usecase.LookupBarcode(c.Request.Context(), barcodeValue)
	if err != nil {
		switch err {
		case domain.ErrBarcodeNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	// Return the medicine details
	c.JSON(http.StatusOK, medicine)
}

// AddBarcodes handles POST /api/barcodes
func (h *BarcodeHandler) AddBarcodes(c *gin.Context) {
	var input domain.Barcode
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Add barcodes using the usecase
	if err := h.usecase.AddBarcodes(c.Request.Context(), input); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Barcodes added successfully"})
}

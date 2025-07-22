package http

import (
	"errors"
	"net/http"
	"strconv"

	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/usecase"
	"pharmacy-management-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// SaleHandler handles sale-related HTTP requests
type SaleHandler struct {
	usecase   usecase.SaleUsecase
	validator *validator.Validate
}

// NewSaleHandler creates a new SaleHandler
func NewSaleHandler(usecase usecase.SaleUsecase, validator *validator.Validate) *SaleHandler {
	return &SaleHandler{usecase, validator}
}

// SearchMedicines handles GET /api/medicines/search
func (h *SaleHandler) SearchMedicines(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("query parameter 'q' is required"))
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	variants, err := h.usecase.SearchMedicines(c.Request.Context(), role.(string), pharmacyID, query)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, variants)
}

// AddToCart handles POST /api/cart
func (h *SaleHandler) AddToCart(c *gin.Context) {
	var input domain.CreateCartInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	role, _ := c.Get("role")
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	if err := h.usecase.AddToCart(c.Request.Context(), role.(string), userID, pharmacyID, input); err != nil {
		switch err {
		case domain.ErrVariantNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		case domain.ErrInsufficientStock:
			utils.ErrorResponse(c, http.StatusBadRequest, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Item added to cart"})
}

// GetCart handles GET /api/cart
func (h *SaleHandler) GetCart(c *gin.Context) {
	role, _ := c.Get("role")
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid user ID"))
		return
	}
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, err := uuid.Parse(pharmacyIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid pharmacy ID"))
		return
	}

	cart, err := h.usecase.GetCart(c.Request.Context(), role.(string), userID, pharmacyID)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, cart)
}

// RemoveFromCart handles DELETE /api/cart/:item_id
func (h *SaleHandler) RemoveFromCart(c *gin.Context) {
	itemIDStr := c.Param("item_id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid cart item ID"))
		return
	}

	role, _ := c.Get("role")
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	if err := h.usecase.RemoveFromCart(c.Request.Context(), role.(string), userID, pharmacyID, itemID); err != nil {
		switch err {
		case domain.ErrCartItemNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
}

// ConfirmSale handles POST /api/sales
func (h *SaleHandler) ConfirmSale(c *gin.Context) {
	role, _ := c.Get("role")
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	sale, err := h.usecase.ConfirmSale(c.Request.Context(), role.(string), userID, pharmacyID)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		case domain.ErrInsufficientStock:
			utils.ErrorResponse(c, http.StatusBadRequest, err)
		case domain.ErrSaleNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Sale confirmed", "sale_id": sale.ID})
}

// GetSales handles GET /api/sales
func (h *SaleHandler) GetSales(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid limit"))
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid offset"))
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, err := uuid.Parse(pharmacyIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid pharmacy ID"))
		return
	}

	sales, err := h.usecase.GetSales(c.Request.Context(), role.(string), pharmacyID, limit, offset)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, sales)
}

// GetReceipt handles GET /api/sales/:id/receipt
func (h *SaleHandler) GetReceipt(c *gin.Context) {
	saleIDStr := c.Param("id")
	saleID, err := uuid.Parse(saleIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid sale ID"))
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, _ := uuid.Parse(pharmacyIDStr.(string))

	receipt, err := h.usecase.GetReceipt(c.Request.Context(), role.(string), pharmacyID, saleID)
	if err != nil {
		switch err {
		case domain.ErrSaleNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	// Map the Receipt data to the flattened ReceiptResponse struct
	response := domain.ReceiptResponse{
		ID:         receipt.ID,
		SaleID:     receipt.SaleID,
		Items:      receipt.Content.Items,
		PharmacyID: receipt.Content.PharmacyID,
		SaleDate:   receipt.Content.SaleDate,
		TotalPrice: receipt.Content.TotalPrice,
		CreatedAt:  receipt.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

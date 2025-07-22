package http

import (
	"errors"
	"net/http"
	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/usecase"
	"pharmacy-management-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	usecase   usecase.OrderUsecase
	validator *validator.Validate
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(usecase usecase.OrderUsecase, validator *validator.Validate) *OrderHandler {
	return &OrderHandler{usecase, validator}
}

// ListOrders handles GET /api/orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
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

	orders, err := h.usecase.ListOrders(c.Request.Context(), role.(string), pharmacyID, limit, offset)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrderDetails handles GET /api/orders/:id
func (h *OrderHandler) GetOrderDetails(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid order ID"))
		return
	}

	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, err := uuid.Parse(pharmacyIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid pharmacy ID"))
		return
	}

	details, err := h.usecase.GetOrderDetails(c.Request.Context(), role.(string), pharmacyID, orderID)
	if err != nil {
		switch err {
		case domain.ErrOrderNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, details)
}

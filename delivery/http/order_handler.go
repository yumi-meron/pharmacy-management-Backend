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

// CreateOrder handles POST /api/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req domain.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, err := uuid.Parse(pharmacyIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid pharmacy ID"))
		return
	}
	order, err := h.usecase.CreateOrder(c.Request.Context(), role.(string), pharmacyID, req)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		case domain.ErrPatientNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}
	c.JSON(http.StatusCreated, order)
}

// RequestOTP handles POST /api/orders/:id/request-otp
func (h *OrderHandler) RequestOTP(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid order ID"))
		return
	}
	var req domain.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, err := uuid.Parse(pharmacyIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid pharmacy ID"))
		return
	}
	err = h.usecase.RequestOTP(c.Request.Context(), role.(string), pharmacyID, orderID, req)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		case domain.ErrOrderNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrInvalidPhone:
			utils.ErrorResponse(c, http.StatusBadRequest, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent"})
}

// VerifyOrder handles POST /api/orders/:id/verify
func (h *OrderHandler) VerifyOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid order ID"))
		return
	}
	var req struct {
		OTP string `json:"otp" validate:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	role, _ := c.Get("role")
	pharmacyIDStr, _ := c.Get("pharmacy_id")
	pharmacyID, err := uuid.Parse(pharmacyIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid pharmacy ID"))
		return
	}
	err = h.usecase.VerifyOrder(c.Request.Context(), role.(string), pharmacyID, orderID, req.OTP)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		case domain.ErrOrderNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case domain.ErrInvalidOTP, domain.ErrOrderAlreadyConfirmed:
			utils.ErrorResponse(c, http.StatusBadRequest, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order confirmed"})
}

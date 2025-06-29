package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yumi-meron/pharmacy-management-app/pharmacist-backend/module/usecase"
)

type AuthHandler struct {
	AuthUC usecase.AuthUsecase
}

func NewAuthHandler(r *gin.Engine, uc usecase.AuthUsecase) {
	handler := &AuthHandler{uc}

	r.POST("/login", handler.Login)
}

type loginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.AuthUC.Login(c.Request.Context(), req.PhoneNumber, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

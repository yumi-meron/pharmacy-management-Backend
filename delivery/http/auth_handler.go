package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pharmacist-backend/domain"
	"github.com/pharmacist-backend/usecase"
)

type AuthHandler struct {
	AuthUC usecase.AuthUsecase
}

func NewAuthHandler(r *gin.Engine, uc usecase.AuthUsecase) {
	handler := &AuthHandler{uc}
	r.POST("/signup", handler.Signup)
	r.POST("/login", handler.Login)
}

type signupRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
	FullName    string `json:"full_name" binding:"required"`
	Role        string `json:"role" binding:"required"` // "owner" or "pharmacist"
	PharmacyID  string `json:"pharmacy_id" binding:"required"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signup data"})
		return
	}

	pharmacyID, _ := uuid.Parse(req.PharmacyID)
	err := h.AuthUC.Signup(c.Request.Context(), usecase.SignupInput{
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
		FullName:    req.FullName,
		Role:        domain.Role(req.Role),
		PharmacyID:  pharmacyID,
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user registered"})
}

type loginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login data"})
		return
	}

	token, err := h.AuthUC.Login(c.Request.Context(), req.PhoneNumber, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

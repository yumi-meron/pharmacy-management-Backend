package http

import (
	"net/http"

	"pharmacist-backend/domain"
	"pharmacist-backend/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	usecase   usecase.AuthUsecase
	validator *validator.Validate
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(usecase usecase.AuthUsecase, validator *validator.Validate) *AuthHandler {
	return &AuthHandler{usecase, validator}
}

// Signup handles POST /signup
func (h *AuthHandler) Signup(c *gin.Context) {
	var input domain.SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user
	if err := h.usecase.Signup(c.Request.Context(), input); err != nil {
		switch err {
		case domain.ErrPhoneNumberTaken:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case domain.ErrInvalidRole:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// Login handles POST /login
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		PhoneNumber string `json:"phone_number" validate:"required,phone"`
		Password    string `json:"password" validate:"required,min=2"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user
	accessToken, refreshToken, err := h.usecase.Login(c.Request.Context(), input.PhoneNumber, input.Password)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// RequestPasswordReset handles POST /password/reset
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var input struct {
		PhoneNumber string `json:"phone_number" validate:"required,phone"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Request reset token
	if err := h.usecase.RequestPasswordReset(c.Request.Context(), input.PhoneNumber); err != nil {
		switch err {
		case domain.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to request password reset"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset code sent"})
}

// ResetPassword handles POST /password/reset/confirm
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Reset password
	if err := h.usecase.ResetPassword(c.Request.Context(), input.Token, input.NewPassword); err != nil {
		switch err {
		case domain.ErrInvalidResetToken:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// RefreshToken handles POST /token/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Refresh tokens
	accessToken, refreshToken, err := h.usecase.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		switch err {
		case domain.ErrInvalidToken:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh token"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

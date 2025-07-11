package http

import (
	"errors"
	"net/http"

	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/usecase"
	"pharmacy-management-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		PhoneNumber string `json:"phone_number" validate:"required,phone"`
		Password    string `json:"password" validate:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Authenticate user
	accessToken, refreshToken, err := h.usecase.Login(c.Request.Context(), input.PhoneNumber, input.Password)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			utils.ErrorResponse(c, http.StatusUnauthorized, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// RequestPasswordReset handles POST /auth/forgot-password
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var input struct {
		PhoneNumber string `json:"phone_number" validate:"required,phone"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Request reset token
	if err := h.usecase.RequestPasswordReset(c.Request.Context(), input.PhoneNumber); err != nil {
		switch err {
		case domain.ErrNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset code sent"})
}

// ResetPassword handles POST /auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Reset password
	if err := h.usecase.ResetPassword(c.Request.Context(), input.Token, input.NewPassword); err != nil {
		switch err {
		case domain.ErrInvalidResetToken:
			utils.ErrorResponse(c, http.StatusBadRequest, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// RefreshToken handles POST /auth/refresh-token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Refresh tokens
	accessToken, refreshToken, err := h.usecase.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		switch err {
		case domain.ErrInvalidToken:
			utils.ErrorResponse(c, http.StatusUnauthorized, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// GetProfile handles GET /api/users/me
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, errors.New("user ID not found in context"))
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid user ID"))
		return
	}

	user, err := h.usecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		switch err {
		case domain.ErrNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              user.ID,
		"phone_number":    user.PhoneNumber,
		"full_name":       user.FullName,
		"role":            user.Role,
		"pharmacy_id":     user.PharmacyID,
		"profile_picture": user.ProfilePicture,
		"created_at":      user.CreatedAt,
		"updated_at":      user.UpdatedAt,
	})
}

// UpdateProfile handles PUT /api/users/me
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var input domain.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Validate input
	if err := h.validator.Struct(input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, errors.New("user ID not found in context"))
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid user ID"))
		return
	}

	if err := h.usecase.UpdateProfile(c.Request.Context(), userID, input); err != nil {
		switch err {
		case domain.ErrPhoneNumberTaken:
			utils.ErrorResponse(c, http.StatusConflict, err)
		case domain.ErrNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

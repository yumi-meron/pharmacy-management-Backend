package utils

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateOTP generates a 6-digit OTP
func GenerateOTP() string {
	otp, _ := GenerateRandomString(12)
	return otp[:6]
}

// ErrorResponse sends a standardized error response
func ErrorResponse(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": err.Error()})
}

// NewValidator creates a new validator with custom validations
func NewValidator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// Simple phone number validation (e.g., starts with +, followed by digits)
		if len(phone) < 10 || phone[0] != '+' {
			return false
		}
		for _, c := range phone[1:] {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	})
	return v
}

// DebugValidator tests custom validations
func DebugValidator(v *validator.Validate) error {
	// Add any validation tests if needed
	return nil
}

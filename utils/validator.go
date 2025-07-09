package utils

import (
	"fmt"
	"regexp"

	"pharmacist-backend/domain"

	"github.com/go-playground/validator/v10"
)

// NewValidator initializes a new validator with custom validations
func NewValidator() *validator.Validate {
	v := validator.New()
	RegisterCustomValidations(v)
	return v
}

// RegisterCustomValidations registers custom validation rules
func RegisterCustomValidations(v *validator.Validate) {
	// Phone number validation (e.g., +251911000000)
	v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		matched, _ := regexp.MatchString(`^\+\d{10,15}$`, phone)
		return matched
	})

	// Role validation
	v.RegisterValidation("role", func(fl validator.FieldLevel) bool {
		role := fl.Field().String()
		return role == string(domain.RolePharmacist) || role == string(domain.RoleOwner) || role == string(domain.RoleAdmin)
	})

	// URL validation for profile_picture and picture
	v.RegisterValidation("url", func(fl validator.FieldLevel) bool {
		url := fl.Field().String()
		if url == "" {
			return true // Allow empty URLs
		}
		matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, url)
		return matched
	})
}

// DebugValidator tests if custom validations are registered
func DebugValidator(v *validator.Validate) error {
	type testStruct struct {
		Phone string `validate:"phone"`
		Role  string `validate:"role"`
		URL   string `validate:"url"`
	}
	testInput := testStruct{
		Phone: "+251911000000",
		Role:  string(domain.RolePharmacist),
		URL:   "https://example.com/profile.jpg",
	}
	if err := v.Struct(testInput); err != nil {
		return fmt.Errorf("validator debug failed: %w", err)
	}
	return nil
}

package domain

import "errors"

var (
	ErrPhoneNumberTaken    = errors.New("phone number already taken")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidResetToken   = errors.New("invalid or expired reset token")

	ErrInvalidInput      = errors.New("invalid input")
	ErrInternalServer    = errors.New("internal server error")
	ErrNotFound          = errors.New("not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidOTP        = errors.New("invalid OTP")
	ErrInvalidRole       = errors.New("invalid role")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrInvalidToken      = errors.New("invalid or expired token")
	ErrPharmacyNotFound  = errors.New("pharmacy is not found")

	ErrMedicineNotFound    = errors.New("medicine not found")
	ErrVariantNotFound     = errors.New("medicine variant not found")
	ErrInvalidPharmacy     = errors.New("invalid pharmacy")
	ErrBarcodeTaken        = errors.New("barcode already taken")
	ErrMedicineHasVariants = errors.New("medicine has variants and cannot be deleted")
)

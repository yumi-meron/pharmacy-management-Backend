package domain

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInternalServer     = errors.New("internal server error")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrNotFound           = errors.New("not found")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPhoneNumberTaken   = errors.New("phone number already in use")
	ErrInvalidOTP         = errors.New("invalid OTP")
)

package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role defines the possible user roles
type Role string

const (
	RolePharmacist Role = "pharmacist"
	RoleOwner      Role = "owner"
	RoleAdmin      Role = "admin"
)

// User represents a user entity in the system
type User struct {
	ID             uuid.UUID `json:"id" validate:"required"`
	PhoneNumber    string    `json:"phone_number" validate:"required,phone" `
	Password       string    `json:"password" validate:"required,min=2"`
	FullName       string    `json:"full_name" validate:"required,min=2,max=100"`
	Role           Role      `json:"role" validate:"required,role"`
	PharmacyID     uuid.UUID `json:"pharmacy_id" validate:"required"`
	ProfilePicture string    `json:"profile_picture" validate:"omitempty,url"` // URL to user profile image
	CreatedAt      time.Time `json:"created_at" validate:"required"`
	UpdatedAt      time.Time `json:"updated_at" validate:"required"`
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	PhoneNumber    string `json:"phone_number" validate:"required,phone"`
	Password       string `json:"password" validate:"required,min=8"`
	FullName       string `json:"full_name" validate:"required"`
	Role           Role   `json:"role" validate:"required,oneof=owner pharmacist"`
	PharmacyID     uuid.UUID
	ProfilePicture string `json:"profile_picture" validate:"omitempty,url"`
}

// UpdateProfileInput represents input for updating a user profile
type UpdateProfileInput struct {
	FullName       string `json:"full_name" validate:"required"`
	PhoneNumber    string `json:"phone_number" validate:"required,phone"`
	Password       string `json:"password" validate:"omitempty,min=8"`
	ProfilePicture string `json:"profile_picture" validate:"omitempty,url"`
}

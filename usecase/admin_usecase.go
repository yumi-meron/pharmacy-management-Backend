package usecase

import (
	"context"
	"time"

	"pharmacist-backend/domain"
	"pharmacist-backend/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AdminUsecase defines the interface for admin-related business logic
type AdminUsecase interface {
	CreateOwner(ctx context.Context, input domain.SignupInput) error
}

// adminUsecase implements AdminUsecase
type adminUsecase struct {
	repo repository.AdminRepository
}

// NewAdminUsecase creates a new AdminUsecase
func NewAdminUsecase(repo repository.AdminRepository) AdminUsecase {
	return &adminUsecase{repo: repo}
}

// CreateOwner creates a user with the owner role
func (u *adminUsecase) CreateOwner(ctx context.Context, input domain.SignupInput) error {
	// Validate role
	if input.Role != domain.RoleOwner {
		return domain.ErrInvalidRole
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create user entity
	user := domain.User{
		ID:             uuid.New(),
		PhoneNumber:    input.PhoneNumber,
		Password:       string(hashedPassword),
		FullName:       input.FullName,
		Role:           input.Role,
		PharmacyID:     input.PharmacyID,
		ProfilePicture: input.ProfilePicture,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save user to database
	return u.repo.CreateOwner(ctx, user)
}

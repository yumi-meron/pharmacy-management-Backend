package usecase

import (
	"context"
	"time"

	"pharmacist-backend/domain"
	"pharmacist-backend/repository"

	"github.com/google/uuid"
)

// PharmacyUsecase defines the interface for pharmacy-related business logic
type PharmacyUsecase interface {
	Create(ctx context.Context, input domain.Pharmacy) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Pharmacy, error)
	GetAll(ctx context.Context, offset, limit int) ([]domain.Pharmacy, error)
	Update(ctx context.Context, input domain.Pharmacy) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// pharmacyUsecase implements PharmacyUsecase
type pharmacyUsecase struct {
	repo repository.PharmacyRepository
}

// NewPharmacyUsecase creates a new PharmacyUsecase
func NewPharmacyUsecase(repo repository.PharmacyRepository) PharmacyUsecase {
	return &pharmacyUsecase{repo: repo}
}

// Create creates a new pharmacy
func (u *pharmacyUsecase) Create(ctx context.Context, input domain.Pharmacy) error {
	// Set timestamps
	input.ID = uuid.New()
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	return u.repo.Create(ctx, input)
}

// GetByID retrieves a pharmacy by ID
func (u *pharmacyUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pharmacy, error) {
	return u.repo.GetByID(ctx, id)
}

// GetAll retrieves all pharmacies with pagination
func (u *pharmacyUsecase) GetAll(ctx context.Context, offset, limit int) ([]domain.Pharmacy, error) {
	return u.repo.GetAll(ctx, offset, limit)
}

// Update updates a pharmacy's details
func (u *pharmacyUsecase) Update(ctx context.Context, input domain.Pharmacy) error {
	// Update timestamp
	input.UpdatedAt = time.Now()

	return u.repo.Update(ctx, input)
}

// Delete deletes a pharmacy by ID
func (u *pharmacyUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if pharmacy exists
	if _, err := u.repo.GetByID(ctx, id); err != nil {
		return err
	}

	return u.repo.Delete(ctx, id)
}

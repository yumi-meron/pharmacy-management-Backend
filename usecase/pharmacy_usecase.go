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
	Create(ctx context.Context, callerRole string, input domain.Pharmacy) error
	GetAll(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID) ([]domain.Pharmacy, error)
	GetByID(ctx context.Context, callerRole string, callerPharmacyID, id uuid.UUID) (*domain.Pharmacy, error)
	Update(ctx context.Context, callerRole string, callerPharmacyID, id uuid.UUID, input domain.Pharmacy) error
	Delete(ctx context.Context, callerRole string, id uuid.UUID) error
}

// pharmacyUsecase implements PharmacyUsecase
type pharmacyUsecase struct {
	repo repository.PharmacyRepository
}

// NewPharmacyUsecase creates a new PharmacyUsecase
func NewPharmacyUsecase(repo repository.PharmacyRepository) PharmacyUsecase {
	return &pharmacyUsecase{repo}
}

// Create creates a new pharmacy (admin-only)
func (u *pharmacyUsecase) Create(ctx context.Context, callerRole string, input domain.Pharmacy) error {
	if callerRole != string(domain.RoleAdmin) {
		return domain.ErrUnauthorized
	}

	input.ID = uuid.New()
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	return u.repo.Create(ctx, input)
}

// GetAll retrieves pharmacies based on role
func (u *pharmacyUsecase) GetAll(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID) ([]domain.Pharmacy, error) {
	if callerRole == string(domain.RoleAdmin) {
		return u.repo.GetAll(ctx)
	}
	pharmacy, err := u.repo.GetByID(ctx, callerPharmacyID)
	if err != nil {
		return nil, err
	}
	return []domain.Pharmacy{*pharmacy}, nil
}

// GetByID retrieves a pharmacy with role-based restrictions
func (u *pharmacyUsecase) GetByID(ctx context.Context, callerRole string, callerPharmacyID, id uuid.UUID) (*domain.Pharmacy, error) {
	if callerRole != string(domain.RoleAdmin) && callerPharmacyID != id {
		return nil, domain.ErrUnauthorized
	}
	return u.repo.GetByID(ctx, id)
}

// Update updates a pharmacy with role-based restrictions
func (u *pharmacyUsecase) Update(ctx context.Context, callerRole string, callerPharmacyID, id uuid.UUID, input domain.Pharmacy) error {
	if callerRole != string(domain.RoleAdmin) && callerPharmacyID != id {
		return domain.ErrUnauthorized
	}
	if callerRole == string(domain.RolePharmacist) {
		return domain.ErrUnauthorized
	}

	input.ID = id
	input.UpdatedAt = time.Now()
	return u.repo.Update(ctx, input)
}

// Delete deletes a pharmacy (admin-only)
func (u *pharmacyUsecase) Delete(ctx context.Context, callerRole string, id uuid.UUID) error {
	if callerRole != string(domain.RoleAdmin) {
		return domain.ErrUnauthorized
	}
	return u.repo.Delete(ctx, id)
}

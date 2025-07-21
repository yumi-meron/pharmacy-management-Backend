package usecase

import (
	"context"
	"time"

	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserUsecase defines the interface for user-related business logic
type UserUsecase interface {
	CreateOwner(ctx context.Context, callerRole string, callerID uuid.UUID, input domain.CreateUserInput) error
	CreatePharmacist(ctx context.Context, callerRole string, callerID uuid.UUID, input domain.CreateUserInput) error
	ListPharmacists(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID) ([]domain.User, error)
}

// userUsecase implements UserUsecase
type userUsecase struct {
	repo repository.AuthRepository
}

// NewUserUsecase creates a new UserUsecase
func NewUserUsecase(repo repository.AuthRepository) UserUsecase {
	return &userUsecase{repo}
}

// CreateOwner creates a new owner (admin-only)
func (u *userUsecase) CreateOwner(ctx context.Context, callerRole string, callerID uuid.UUID, input domain.CreateUserInput) error {
	if callerRole != string(domain.RoleAdmin) {
		return domain.ErrUnauthorized
	}

	if input.Role != domain.RoleOwner {
		return domain.ErrInvalidRole
	}

	// Check if phone number is taken
	if existingUser, _ := u.repo.GetByPhone(ctx, input.PhoneNumber); existingUser != nil {
		return domain.ErrPhoneNumberTaken
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

	return u.repo.Create(ctx, user)
}

// CreatePharmacist creates a new pharmacist (admin or owner)
func (u *userUsecase) CreatePharmacist(ctx context.Context, callerRole string, callerID uuid.UUID, input domain.CreateUserInput) error {
	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) {
		return domain.ErrUnauthorized
	}

	if input.Role != domain.RolePharmacist {
		return domain.ErrInvalidRole
	}

	// If caller is Owner, ensure pharmacy_id matches their own
	if callerRole == string(domain.RoleOwner) {
		caller, err := u.repo.GetByID(ctx, callerID)
		if err != nil {
			return err
		}
		if caller.PharmacyID != input.PharmacyID {
			return domain.ErrUnauthorized
		}
	}

	// Check if phone number is taken
	if existingUser, _ := u.repo.GetByPhone(ctx, input.PhoneNumber); existingUser != nil {
		return domain.ErrPhoneNumberTaken
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

	return u.repo.Create(ctx, user)
}
func (u *userUsecase) ListPharmacists(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID) ([]domain.User, error) {
	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) {
		return nil, domain.ErrUnauthorized
	}

	var pharmacists []domain.User
	var err error

	if callerRole == string(domain.RoleAdmin) {
		// Admins can see all pharmacists
		pharmacists, err = u.repo.GetPharmacists(ctx, nil)
	} else {
		// Owners can only see pharmacists in their pharmacy
		pharmacists, err = u.repo.GetPharmacists(ctx, &callerPharmacyID)
	}

	if err != nil {
		return nil, err
	}
	return pharmacists, nil
}

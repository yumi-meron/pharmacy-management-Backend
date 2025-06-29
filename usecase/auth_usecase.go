package usecase

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/pharmacist-backend/domain"
	"github.com/pharmacist-backend/repository"
)

type AuthUsecase interface {
	Signup(ctx context.Context, input SignupInput) error
	Login(ctx context.Context, phone, password string) (string, error)
}

type authUsecase struct {
	repo repository.AuthRepository
}

func NewAuthUsecase(repo repository.AuthRepository) AuthUsecase {
	return &authUsecase{repo}
}

type SignupInput struct {
	PhoneNumber string
	Password    string
	FullName    string
	Role        domain.Role
	PharmacyID  uuid.UUID
}

func (uc *authUsecase) Signup(ctx context.Context, input SignupInput) error {
	existing, _ := uc.repo.GetByPhone(ctx, input.PhoneNumber)
	if existing != nil {
		return errors.New("phone number already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &domain.User{
		ID:           uuid.New(),
		PhoneNumber:  input.PhoneNumber,
		PasswordHash: string(hash),
		FullName:     input.FullName,
		Role:         input.Role,
		PharmacyID:   input.PharmacyID,
	}
	return uc.repo.Create(ctx, user)
}

func (uc *authUsecase) Login(ctx context.Context, phone, password string) (string, error) {
	user, err := uc.repo.GetByPhone(ctx, phone)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     user.ID.String(),
		"role":        user.Role,
		"pharmacy_id": user.PharmacyID.String(),
		"exp":         time.Now().Add(72 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

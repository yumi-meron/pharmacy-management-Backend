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
	"github.com/pharmacist-backend/infrastructure"
	"github.com/pharmacist-backend/repository"
	"github.com/pharmacist-backend/utils"
)

type AuthUsecase interface {
	Signup(ctx context.Context, input SignupInput) error
	Login(ctx context.Context, phone, password string) (string, error)
	RequestLoginOTP(ctx context.Context, phone string) error
	VerifyLoginOTP(ctx context.Context, phone, otp string) (string, error)
}

type authUsecase struct {
	repo          repository.AuthRepository
	twilioService *infrastructure.TwilioService
}

func NewAuthUsecase(repo repository.AuthRepository, twilioService *infrastructure.TwilioService) AuthUsecase {
	return &authUsecase{repo, twilioService}
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
		return domain.ErrPhoneNumberTaken
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
		return "", domain.ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", domain.ErrInvalidCredentials
	}

	otp := utils.GenerateOTP()
	expiresAt := time.Now().Add(5 * time.Minute)
	if err := uc.repo.SaveOTP(ctx, phone, otp, expiresAt); err != nil {
		return "", err
	}

	if err := uc.twilioService.SendOTP(ctx, phone, otp); err != nil {
		return "", err
	}

	return "", errors.New("OTP sent for verification")
}

func (uc *authUsecase) RequestLoginOTP(ctx context.Context, phone string) error {
	otp := utils.GenerateOTP()
	expiresAt := time.Now().Add(5 * time.Minute)
	if err := uc.repo.SaveOTP(ctx, phone, otp, expiresAt); err != nil {
		return err
	}
	return uc.twilioService.SendOTP(ctx, phone, otp)
}

func (uc *authUsecase) VerifyLoginOTP(ctx context.Context, phone, otp string) (string, error) {
	valid, err := uc.repo.VerifyOTP(ctx, phone, otp)
	if err != nil || !valid {
		return "", domain.ErrInvalidOTP
	}

	user, err := uc.repo.GetByPhone(ctx, phone)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     user.ID.String(),
		"role":        user.Role,
		"pharmacy_id": user.PharmacyID.String(),
		"exp":         time.Now().Add(72 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

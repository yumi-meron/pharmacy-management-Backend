package usecase

import (
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yumi-meron/pharmacy-management-app/pharmacist-backend/module/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase interface {
	Login(ctx context.Context, phone string, password string) (string, error)
}

type authUsecase struct {
	repo repository.AuthRepository
}

func NewAuthUsecase(repo repository.AuthRepository) AuthUsecase {
	return &authUsecase{repo}
}

func (uc *authUsecase) Login(ctx context.Context, phone, password string) (string, error) {
	user, err := uc.repo.GetByPhone(ctx, phone)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", err
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     user.ID.String(),
		"role":        user.Role,
		"pharmacy_id": user.PharmacyID.String(),
		"exp":         time.Now().Add(time.Hour * 72).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}

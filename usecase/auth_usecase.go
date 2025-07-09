package usecase

import (
	"context"
	"time"

	"pharmacist-backend/config"
	"pharmacist-backend/domain"
	"pharmacist-backend/infrastructure"
	"pharmacist-backend/repository"
	"pharmacist-backend/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthUsecase defines the interface for authentication-related business logic
type AuthUsecase interface {
	Signup(ctx context.Context, input domain.SignupInput) error
	Login(ctx context.Context, phoneNumber, password string) (string, string, error)
	RequestPasswordReset(ctx context.Context, phoneNumber string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
}

// authUsecase implements AuthUsecase
type authUsecase struct {
	repo   repository.AuthRepository
	twilio *infrastructure.TwilioService
	cfg    *config.Config
}

// NewAuthUsecase creates a new AuthUsecase
func NewAuthUsecase(repo repository.AuthRepository, twilio *infrastructure.TwilioService, cfg *config.Config) AuthUsecase {
	return &authUsecase{repo, twilio, cfg}
}

// Signup creates a new user
func (u *authUsecase) Signup(ctx context.Context, input domain.SignupInput) error {
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

	// Save user to database
	return u.repo.Create(ctx, user)
}

// Login authenticates a user and returns access and refresh tokens
func (u *authUsecase) Login(ctx context.Context, phoneNumber, password string) (string, string, error) {
	user, err := u.repo.GetByPhone(ctx, phoneNumber)
	if err == domain.ErrNotFound {
		return "", "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return "", "", err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", domain.ErrInvalidCredentials
	}

	// Generate access token
	accessToken, err := u.generateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := u.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	// Save refresh token
	if err := u.repo.SaveRefreshToken(ctx, user.ID, refreshToken, time.Now().Add(7*24*time.Hour)); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// RequestPasswordReset sends a reset token via SMS
func (u *authUsecase) RequestPasswordReset(ctx context.Context, phoneNumber string) error {
	user, err := u.repo.GetByPhone(ctx, phoneNumber)
	if err == domain.ErrNotFound {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}

	// Generate OTP
	otp := utils.GenerateOTP()
	expiry := time.Now().Add(15 * time.Minute)

	// Save reset token
	if err := u.repo.SaveResetToken(ctx, user.ID, otp, expiry); err != nil {
		return err
	}

	// Send OTP via SMS
	message := "Your password reset code is: " + otp
	return u.twilio.SendSMS(user.PhoneNumber, message)
}

// ResetPassword updates the user's password using a reset token
func (u *authUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	userID, err := u.repo.GetResetToken(ctx, token)
	if err != nil {
		return err
	}

	// Retrieve user
	user, err := u.repo.GetByID(ctx, *userID)
	if err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()
	if err := u.repo.Update(ctx, *user); err != nil {
		return err
	}

	// Delete reset token
	return u.repo.DeleteResetToken(ctx, token)
}

// RefreshToken generates new access and refresh tokens
func (u *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	userID, err := u.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	user, err := u.repo.GetByID(ctx, *userID)
	if err != nil {
		return "", "", err
	}

	// Generate new access token
	accessToken, err := u.generateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", err
	}

	// Generate new refresh token
	newRefreshToken, err := u.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	// Delete old refresh token
	if err := u.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return "", "", err
	}

	// Save new refresh token
	if err := u.repo.SaveRefreshToken(ctx, user.ID, newRefreshToken, time.Now().Add(7*24*time.Hour)); err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// generateAccessToken creates a JWT access token
func (u *authUsecase) generateAccessToken(userID uuid.UUID, role domain.Role) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.cfg.JWTSecret))
}

// generateRefreshToken creates a refresh token
func (u *authUsecase) generateRefreshToken(userID uuid.UUID) (string, error) {
	token := utils.GenerateRandomString(32)
	return token, nil
}

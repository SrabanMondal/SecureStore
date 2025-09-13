package services

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/SrabanMondal/SecureStore/internal/utils"
	"github.com/SrabanMondal/SecureStore/internal/models"
	"github.com/SrabanMondal/SecureStore/internal/repository"
)

type AuthService struct {
	UserRepo   *repositories.UserRepository
	JWTSecret  string
	JWTExpiry  time.Duration
}

func NewAuthService(userRepo *repositories.UserRepository, secret string, expiry time.Duration) *AuthService {
	return &AuthService{
		UserRepo:  userRepo,
		JWTSecret: secret,
		JWTExpiry: expiry,
	}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (*models.User, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
	}

	err = s.UserRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	token, err := utils.GenerateJWT(user.ID, s.JWTSecret, s.JWTExpiry)
	if err != nil {
		return "", err
	}

	return token, nil
}

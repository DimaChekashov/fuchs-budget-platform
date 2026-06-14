package service

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      *repository.UserRepository
	jwtSecret string
	expiry    time.Duration
}

func NewAuthService(repo *repository.UserRepository, jwtSecret string, expiryHours int) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: jwtSecret,
		expiry:    time.Duration(expiryHours) * time.Hour,
	}
}

type LoginInput struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken string `json:"access_token"`
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &TokenPair{AccessToken: token}, nil
}

func (s *AuthService) generateToken(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(s.expiry).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}

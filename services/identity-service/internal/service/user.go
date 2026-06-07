package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/domain"
	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

type RegisterInput struct {
	Email    string
	Password string
}

func (s *UserService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	existing, _ := s.repo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Email:    input.Email,
		Password: string(hashed),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

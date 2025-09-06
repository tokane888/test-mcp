package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokane888/go-repository-template/services/api/internal/domain"
	"github.com/tokane888/go-repository-template/services/api/internal/dto/request"
	"github.com/tokane888/go-repository-template/services/api/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserUseCase interface {
	CreateUser(ctx context.Context, req *request.CreateUser) (*domain.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, int, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type userUseCase struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewUserUseCase(userRepo repository.UserRepository, logger *zap.Logger) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (uc *userUseCase) CreateUser(ctx context.Context, req *request.CreateUser) (*domain.User, error) {
	// Check if user already exists
	exists, err := uc.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// Create domain user entity
	user, err := domain.NewUser(req.Email, req.Username, req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user entity: %w", err)
	}

	// Save to repository
	if err := uc.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

func (uc *userUseCase) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	users, total, err := uc.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

func (uc *userUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Find user
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Mark as deleted
	user.Delete()

	// Update in repository
	if err := uc.userRepo.Update(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

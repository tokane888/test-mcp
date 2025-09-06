package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokane888/go-repository-template/services/api/internal/domain"
	"github.com/tokane888/go-repository-template/services/api/internal/repository"
	"go.uber.org/zap"
)

type userRepositoryImpl struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewUserRepository(db *sql.DB, logger *zap.Logger) repository.UserRepository {
	return &userRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID(),
		user.Email(),
		user.Username(),
		user.PasswordHash(),
		user.CreatedAt(),
		user.UpdatedAt(),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return repository.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	var (
		userID       uuid.UUID
		email        string
		username     string
		passwordHash string
		createdAt    sql.NullTime
		updatedAt    sql.NullTime
		deletedAt    sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&userID,
		&email,
		&username,
		&passwordHash,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	user := domain.ReconstructUser(
		userID,
		email,
		username,
		passwordHash,
		createdAt.Time,
		updatedAt.Time,
		getTimePtr(deletedAt),
	)

	return user, nil
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	var (
		userID       uuid.UUID
		userEmail    string
		username     string
		passwordHash string
		createdAt    sql.NullTime
		updatedAt    sql.NullTime
		deletedAt    sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&userID,
		&userEmail,
		&username,
		&passwordHash,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	user := domain.ReconstructUser(
		userID,
		userEmail,
		username,
		passwordHash,
		createdAt.Time,
		updatedAt.Time,
		getTimePtr(deletedAt),
	)

	return user, nil
}

func (r *userRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	// Count total users
	var total int
	countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users
	query := `
		SELECT id, email, username, password_hash, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error("failed to close rows", zap.Error(closeErr))
		}
	}()

	var users []*domain.User
	for rows.Next() {
		var (
			userID       uuid.UUID
			email        string
			username     string
			passwordHash string
			createdAt    sql.NullTime
			updatedAt    sql.NullTime
			deletedAt    sql.NullTime
		)

		if scanErr := rows.Scan(
			&userID,
			&email,
			&username,
			&passwordHash,
			&createdAt,
			&updatedAt,
			&deletedAt,
		); scanErr != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", scanErr)
		}

		user := domain.ReconstructUser(
			userID,
			email,
			username,
			passwordHash,
			createdAt.Time,
			updatedAt.Time,
			getTimePtr(deletedAt),
		)

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return users, total, nil
}

func (r *userRepositoryImpl) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, username = $2, updated_at = $3, deleted_at = $4
		WHERE id = $5`

	result, err := r.db.ExecContext(ctx, query,
		user.Email(),
		user.Username(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.ID(),
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

func (r *userRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)"
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}
	return exists, nil
}

// Helper function to detect PostgreSQL unique constraint violation
func isUniqueViolation(err error) bool {
	return err != nil && (err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" ||
		err.Error() == "pq: duplicate key value violates unique constraint \"users_email_unique_not_deleted\"" ||
		containsString(err.Error(), "duplicate key value violates unique constraint"))
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && containsString(s[1:], substr)
}

func getTimePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

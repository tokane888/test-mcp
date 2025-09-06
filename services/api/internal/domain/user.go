package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters")
	ErrUsernameTooShort      = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong       = errors.New("username must be at most 100 characters")
	ErrInvalidPasswordFormat = errors.New("password must contain at least one letter and one number")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type User struct {
	id           uuid.UUID
	email        string
	username     string
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// NewUser creates a new User entity with validation
func NewUser(email, username, password string) (*User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validateUsername(username); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		id:           uuid.New(),
		email:        email,
		username:     username,
		passwordHash: hashedPassword,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ReconstructUser reconstructs a User entity from persistence
func ReconstructUser(
	id uuid.UUID,
	email string,
	username string,
	passwordHash string,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *User {
	return &User{
		id:           id,
		email:        email,
		username:     username,
		passwordHash: passwordHash,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}
}

// Getters
func (u *User) ID() uuid.UUID         { return u.id }
func (u *User) Email() string         { return u.email }
func (u *User) Username() string      { return u.username }
func (u *User) PasswordHash() string  { return u.passwordHash }
func (u *User) CreatedAt() time.Time  { return u.createdAt }
func (u *User) UpdatedAt() time.Time  { return u.updatedAt }
func (u *User) DeletedAt() *time.Time { return u.deletedAt }

// Business methods
func (u *User) Delete() {
	now := time.Now()
	u.deletedAt = &now
	u.updatedAt = now
}

func (u *User) IsDeleted() bool {
	return u.deletedAt != nil
}

func (u *User) UpdateEmail(email string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	u.email = email
	u.updatedAt = time.Now()
	return nil
}

func (u *User) UpdateUsername(username string) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	u.username = username
	u.updatedAt = time.Now()
	return nil
}

func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(password))
	return err == nil
}

// Validation functions
func validateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

func validateUsername(username string) error {
	if len(username) < 3 {
		return ErrUsernameTooShort
	}
	if len(username) > 100 {
		return ErrUsernameTooLong
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	// Check if password contains at least one letter and one number
	hasLetter := false
	hasNumber := false
	for _, char := range password {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
		if hasLetter && hasNumber {
			break
		}
	}
	if !hasLetter || !hasNumber {
		return ErrInvalidPasswordFormat
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

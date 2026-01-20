package auth

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mr1hm/go-chat-moderator/internal/shared/sqlite"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrEmailExists    = errors.New("email already exists")
	ErrUsernameExists = errors.New("username already exists")
)

type UserRepository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
}

type sqliteUserRepo struct{}

func NewUserRepository() UserRepository {
	return &sqliteUserRepo{}
}

func (r *sqliteUserRepo) Create(user *User) error {
	user.ID = uuid.New().String()

	_, err := sqlite.DB.Exec(
		`INSERT INTO users (id, email, password_hash, username) VALUES (?, ?, ?, ?)`,
		user.ID, user.Email, user.PasswordHash, user.Username,
	)
	if err != nil {
		// Check for unique constraint violations
		if isUniqueViolation(err, "email") {
			return ErrEmailExists
		}
		if isUniqueViolation(err, "username") {
			return ErrUsernameExists
		}
		return err
	}

	return nil
}

func (r *sqliteUserRepo) FindByEmail(email string) (*User, error) {
	user := &User{}
	err := sqlite.DB.QueryRow(
		`SELECT id, email, password_hash, username, created_at, updated_at FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	return user, err
}

func (r *sqliteUserRepo) FindByID(id string) (*User, error) {
	user := &User{}
	err := sqlite.DB.QueryRow(
		`SELECT id, email, password_hash, username, created_at, updated_at FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return user, err
}

func isUniqueViolation(err error, field string) bool {
	// SQLite unique constraint error contains "UNIQUE constraint failed"
	return err != nil && strings.Contains(err.Error(), "UNIQUE") && strings.Contains(err.Error(), field)
}

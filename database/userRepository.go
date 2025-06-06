package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
)

var (
	ErrUserNotFound = errors.New("no user found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepo(conn *Connection) *UserRepository {
	return &UserRepository{
		db: conn.db,
	}
}

func (r *UserRepository) Save(ctx context.Context, user domain.User) error {
	query := "INSERT INTO users(id, fullName, email, phone, password, birthDate, registerDate) VALUES ($1, $2, $3, $4, $5, $6, $7);"
	_, err := r.db.ExecContext(
		ctx, query, user.Id, user.FullName, user.Email, user.Phone, user.PasswordHash, user.BirthDate, user.RegisterDate,
	)

	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == "23505" {
			return fmt.Errorf("unique constraint violation - %w", ErrUserAlreadyExists)
		}
		return fmt.Errorf("saving operation failed: %w", err)
	}

	return nil
}

func (r *UserRepository) Get(ctx context.Context, email string) (*domain.User, error) {
	user := domain.User{}
	query := "SELECT * FROM users WHERE email=$1;"
	
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id, &user.FullName, &user.Email, &user.Phone, &user.PasswordHash, &user.BirthDate, &user.RegisterDate,
	); err != nil {
		if err == sql.ErrNoRows {
            return nil, fmt.Errorf("can't find user with email=%s - %w", email, ErrUserNotFound)
        }
		return nil, fmt.Errorf("user retrieve operation failed: %w", err)
	}

	return &user, nil
}

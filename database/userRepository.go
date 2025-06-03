package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
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
		ctx, query, user.Id, user.FullName, user.Email, user.Phone, user.Password, user.BirthDate, user.RegisterDate,
	)

	if err != nil {
		return fmt.Errorf("user saving operation failed: %w", err)
	}

	return nil
}

func (r *UserRepository) Get(ctx context.Context, email string) (*domain.User, error) {
	user := domain.User{}
	query := "SELECT * FROM users WHERE email=$1;"
	
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id, &user.FullName, &user.Email, &user.Phone, &user.Password, &user.BirthDate, &user.RegisterDate,
	); err != nil {
		return nil, fmt.Errorf("user retrieve operation failed: %w", err)
	}

	return &user, nil
}

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

func NewUserRepository(conf *Config) (*UserRepository, error) {
	if err := CreateDBIfNotExists(conf); err != nil {
		return nil, err
	}

	db, err := sql.Open(conf.Driver, conf.GetPgConnString(false))
	if err != nil {
		fmt.Println("invalid connection arguments:", err)
		return nil, err	
	}

	if err := db.Ping(); err != nil {
		fmt.Println("database connection failed:", err)
		return nil, err
	}	
	
	return &UserRepository{
		db: db,
	}, nil
}

func CreateDBIfNotExists(conf *Config) error {
	defaultConn, err := sql.Open(conf.Driver, conf.GetPgConnString(true))
	if err != nil {
		fmt.Println("invalid default connection arguments:", err)
		return err
	}

	defer defaultConn.Close()
	var exists bool

	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", conf.DBName)
	err = defaultConn.QueryRow(query).Scan(&exists)
	if err != nil {
		fmt.Println("database existense check query failed:", err)
		return err
	}

	if !exists {
		_, err := defaultConn.Exec("CREATE DATABASE " + conf.DBName)
		if err != nil {
			fmt.Println("database creation failed:", err)
			return err
		}
	}

	return nil
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

func (r *UserRepository) Delete(ctx context.Context, userId uuid.UUID) error {
	query := "DELETE FROM users WHERE id=$1;"

	result, err := r.db.ExecContext(ctx, query); 
	if err != nil {
		return fmt.Errorf("user remove operation failed: %w", err)
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) Exit() {
	if r.db != nil {
		r.db.Close()
	}
}

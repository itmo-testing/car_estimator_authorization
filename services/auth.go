package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/database"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type IUserProvider interface {
	Get(ctx context.Context, email string) (*domain.User, error)
}

type IUserSaver interface {
	Save(ctx context.Context, user domain.User) error
}

type AuthService struct {
	userProvider IUserProvider
	userSaver IUserSaver
	logger *slog.Logger
}

func NewAuthService (provider IUserProvider, saver IUserSaver, logger *slog.Logger) *AuthService {
	return &AuthService{
		userProvider: provider,
		userSaver: saver,
		logger: logger,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (token string, err error) {
	log := s.logger.With(
		slog.String("operation", "login"),
		slog.String("email", email),
	)

	log.Info("authorization attempt...")

	user, err := s.userProvider.Get(ctx, email)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			s.logger.WarnContext(ctx, "user not found", slog.Any("error", err))
		}
		return "", fmt.Errorf("login error - %w", ErrInvalidCredentials)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return "", fmt.Errorf("login error - %w", ErrInvalidCredentials)
	}
	
	if token, err = CreateJWT(user); err != nil {
		return "", err
	}

	log.Info("successfully logged in!")
	return token, nil
}


func (s *AuthService) Register(ctx context.Context, user domain.User) (userId uuid.UUID, err error) {
	log := s.logger.With(
		slog.String("operation", "register"),
		slog.String("name", user.FullName),
		slog.String("email", user.Email),
	)

	log.Info("proceeding registration...")

	user.Id = uuid.New()
	user.RegisterDate = time.Now()

	user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate password hash", slog.Any("error", err))
		return uuid.UUID{}, err
	}

	if err = s.userSaver.Save(ctx, user); err != nil {
		if errors.Is(err, database.ErrUserAlreadyExists) {
			s.logger.WarnContext(ctx, "user already exists")
		}
		s.logger.ErrorContext(ctx, "failed to register user", slog.Any("error", err))
		return uuid.UUID{}, err
	}

	log.Info("registration complete!")
	return user.Id, nil
}

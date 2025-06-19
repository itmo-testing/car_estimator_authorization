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

type IUserSaver interface {
	Save(ctx context.Context, user domain.User) error
}

type IUserRemover interface {
	Delete(ctx context.Context, userId uuid.UUID) error
}

type RegistrarService struct {
	userSaver IUserSaver
	userRemover IUserRemover
	sessionRemover ISessionRemover
	sessionProvider ISessionProvider
	logger *slog.Logger
}

func NewRegistrarService (
		userSaver IUserSaver,
		userRemover IUserRemover,
		sessionRemover ISessionRemover,
		sessionProvider ISessionProvider,
		logger *slog.Logger) *RegistrarService {
	return &RegistrarService{
		userSaver: userSaver,
		userRemover: userRemover,
		sessionRemover: sessionRemover,
		sessionProvider: sessionProvider,
		logger: logger,
	}
}

func (s *RegistrarService) Register(ctx context.Context, user domain.User) (userId uuid.UUID, err error) {
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

func (s *RegistrarService) Unregister(ctx context.Context, refreshToken string) error {
	log := s.logger.With(
		slog.String("operation", "unregister"),
	)

	log.Info("start account delete procedure...")

	session, err := s.sessionProvider.Get(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, database.ErrSessionNotFound) {
			s.logger.WarnContext(ctx, "no session found", slog.Any("error", err))
		}
		return fmt.Errorf("unregister error - %w", err)
	}

	log = log.With(slog.String("userId", session.UserId.String()))

	if err = s.sessionRemover.DeleteUserSessions(ctx, session.UserId); err != nil {
		return fmt.Errorf("unregister error - %w", err)
	}

	if err = s.userRemover.Delete(ctx, session.UserId); err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			s.logger.WarnContext(ctx, "user not found", slog.Any("error", err))
		}
		return fmt.Errorf("unregister error - %w", err)
	}

	log.Info("successfully unregistered user!")
	return nil
}

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
	ErrSourceChanged = errors.New("source changed")
	ErrAlreadyLoggedIn = errors.New("already logged in")
)

type IUserProvider interface {
	Get(ctx context.Context, email string) (*domain.User, error)
	GetById(ctx context.Context, userId uuid.UUID) (*domain.User, error)
}

type ISessionProvider interface {
	Get(ctx context.Context, token string) (*domain.Session, error)
	GetUserSessions(ctx context.Context, userId uuid.UUID) ([]*domain.Session, error)
}

type ISessionSaver interface {
	Save(ctx context.Context, session *domain.Session, expiresIn time.Duration) (string, error)
}

type ISessionRemover interface {
	Delete(ctx context.Context, token string) error
	DeleteUserSessions(ctx context.Context, userId uuid.UUID) error
}

type AuthService struct {
	userProvider IUserProvider
	sessionProvider ISessionProvider
	sessionSaver ISessionSaver
	sessionRemover ISessionRemover
	logger *slog.Logger
}

func NewAuthService (
		userProvider IUserProvider,
		sessionProvider ISessionProvider, 
		sessionSaver ISessionSaver,
		sessionRemover ISessionRemover,
		logger *slog.Logger) *AuthService {
	return &AuthService{
		userProvider: userProvider,
		sessionProvider: sessionProvider,
		sessionSaver: sessionSaver,
		sessionRemover: sessionRemover,
		logger: logger,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string, source domain.Source) (*domain.TokenPair, *uuid.UUID, error) {
	log := s.logger.With(
		slog.String("operation", "login"),
		slog.String("email", email),
		slog.String("ip address", source.IpAddress),
		slog.String("user agent", source.UserAgent),
	)

	log.Info("authorization attempt...")

	user, err := s.userProvider.Get(ctx, email)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			s.logger.WarnContext(ctx, "user not found", slog.Any("error", err))
		}
		return nil, nil, fmt.Errorf("login error - %w", ErrInvalidCredentials)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return nil, nil, fmt.Errorf("login error - %w", ErrInvalidCredentials)
	}

	userSessions, err := s.sessionProvider.GetUserSessions(ctx, user.Id)
	if err != nil {
		return nil, nil, fmt.Errorf("login error - user sessions search failure: %w", err)
	}

	for _, session := range userSessions {
		if source.IpAddress == session.IpAddress && source.UserAgent == session.UserAgent {
			return nil, nil, fmt.Errorf("login error - %w", ErrAlreadyLoggedIn)
		}
	}
	
	accessToken, err := CreateJWT(user)
	if err != nil {
		return nil, nil, err
	}

	newSession := &domain.Session{
		UserId: user.Id,
		Email: user.Email,
		Source: source,
		CreatedAt: time.Now(),
	}

	refreshToken, err := s.sessionSaver.Save(ctx, newSession, time.Hour * 24 * 30)
	if err != nil {
		return nil, nil, err
	}

	log.Info("successfully logged in!")
	return &domain.TokenPair{
		Access: accessToken,
		Refresh: refreshToken,
	}, &(user.Id), nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	log := s.logger.With(
		slog.String("operation", "logout"),
	)

	log.Info("exiting the system...")

	if err := s.sessionRemover.Delete(ctx, refreshToken); err != nil {
		if errors.Is(err, database.ErrSessionNotFound) {
			s.logger.WarnContext(ctx, "session is not found", slog.Any("error", err))
		}
		return fmt.Errorf("logout error - %w", err)
	}

	log.Info("successfully exited!")
	return nil
}

func (s *AuthService) GetUser(ctx context.Context, userId uuid.UUID) (*domain.UserPublic, error) {
	log := s.logger.With(
		slog.String("operation", "get user"),
		slog.String("User ID", userId.String()),
	)

	log.Info("try to retreive user from database...")

	user, err := s.userProvider.GetById(ctx, userId)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			s.logger.WarnContext(ctx, "user not found", slog.Any("error", err))
		}
		return nil, fmt.Errorf("get user error - %w", err)
	}

	log.Info("successfully retreived user!", 
		slog.Any("user data", user.UserPublic),
	)

	return &user.UserPublic, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string, source domain.Source) (*domain.TokenPair, error) {
	log := s.logger.With(
		slog.String("operation", "refresh"),
		slog.String("ip address", source.IpAddress),
		slog.String("user agent", source.UserAgent),
	)

	log.Info("try to refresh tokens...")

	session, err := s.sessionProvider.Get(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, database.ErrSessionNotFound) {
			s.logger.WarnContext(ctx, "no session found", slog.Any("error", err))
		}
		return nil, fmt.Errorf("refresh error - %w", err)
	}

	if err = s.sessionRemover.Delete(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("refresh error - %w", err)
	}

	if session.IpAddress != source.IpAddress && session.UserAgent != source.UserAgent {
		return nil, ErrSourceChanged
	}

	newRefreshToken, err := s.sessionSaver.Save(ctx, session, time.Hour * 24 * 30)
	if err != nil {
		return nil, err
	}

	accessToken, err := CreateJWT(&domain.User{
		UserPublic: domain.UserPublic{
			Id: session.UserId, 
			Email: session.Email,
		},
	})

	if err != nil {
		return nil, err
	}

	log.Info("successfully refreshed tokens!")
	return &domain.TokenPair{
		Access: accessToken,
		Refresh: newRefreshToken,
	}, nil
}

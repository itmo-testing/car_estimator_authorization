package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
	"github.com/redis/go-redis/v9"
)

var (
	storageName = "sessions:"
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
	refreshTokenLen = 64
	ErrSessionNotFound = errors.New("session not found")
)

type SessionRepository struct {
	client *redis.Client
}

func generateRefreshToken() string {
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, refreshTokenLen)
	for i := 0; i < refreshTokenLen; i++ {
		b[i] = allowed[random.Intn(len(allowed))]
	}
	return string(b)
}

func NewSessionRepository(conf *Config) (*SessionRepository, error) {
	options, err := redis.ParseURL(conf.GetRedisConnString())
	if err != nil {
		return nil, fmt.Errorf("error while creating session repository: %w", err)
	}

	client := redis.NewClient(options)
	ctx := context.Background()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis connection error: %w", err)
	}

	return &SessionRepository{
		client: client,
	}, nil
}

func (r *SessionRepository) Save(ctx context.Context, session *domain.Session, expiresIn time.Duration) (string, error) {
	binary, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("session serialization error: %w", err)
	}

	newToken := generateRefreshToken()
	if err := r.client.Set(ctx, storageName + newToken, binary, expiresIn).Err(); err != nil {
		return "", fmt.Errorf("redis error - session saving failed: %w", err)
	}
	return newToken, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	deleted, err := r.client.Del(ctx, storageName + token).Result(); 
	if err != nil {
		return fmt.Errorf("redis error - session delete failed: %w", err)
	}

	if deleted == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (r *SessionRepository) Get(ctx context.Context, token string) (*domain.Session, error) {
	session := domain.Session{}
	binary, err := r.client.Get(ctx, storageName + token).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf("redis error - retrieve op failed: %w", err)
	}

	if err = json.Unmarshal(binary, &session); err != nil {
		return nil, fmt.Errorf("session deserialization error: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) Exit() {
	if r != nil {
		_ = r.client.Close()
	}
} 

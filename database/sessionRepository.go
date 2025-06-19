package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
	"github.com/redis/go-redis/v9"
)

var (
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
	if err = r.client.Set(ctx, newToken, binary, expiresIn).Err(); err != nil {
		return "", fmt.Errorf("redis error - session saving failed: %w", err)
	}

	if err = r.client.SAdd(ctx, session.UserId.String(), newToken).Err(); err != nil {
		return "", fmt.Errorf("redis error - can't add to user sessions set: %w", err)
	}
	return newToken, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	session, err := r.Get(ctx, token)
	if err != nil {
		return err
	}

	_, err = r.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, token)
        pipe.SRem(ctx, session.UserId.String(), token)
        return nil
	})

	if err != nil {
		return fmt.Errorf("redis error - can't delete user session: %w", err)
	}

	return nil
}

func (r *SessionRepository) DeleteUserSessions(ctx context.Context, userId uuid.UUID) error {
	sessions, err := r.client.SMembers(ctx, userId.String()).Result()
    if err != nil {
        return fmt.Errorf("redis error - user sessions search failed: %w", err)
    }

    if len(sessions) > 0 {
        if err := r.client.Del(ctx, sessions...).Err(); err != nil {
            return err
        }
    }

    return r.client.Del(ctx, userId.String()).Err()
}

func (r *SessionRepository) Get(ctx context.Context, token string) (*domain.Session, error) {
	session := domain.Session{}
	binary, err := r.client.Get(ctx, token).Bytes()
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

func (r *SessionRepository) GetUserSessions(ctx context.Context, userId uuid.UUID) ([]*domain.Session, error) {
	result := make([]*domain.Session, 0)
	sessions, err := r.client.SMembers(ctx, userId.String()).Result()
	if err != nil {
		return nil, fmt.Errorf("redis error - user sessions search failed: %w", err)
	}

	for _, refreshToken := range sessions {
		s, err := r.Get(ctx, refreshToken)
		if err != nil {
			log.Printf("[WARNING] session deleted from its storage but not from user's set. token=%s", refreshToken)
			continue
		}
		result = append(result, s)
	}
	return result, nil
}

func (r *SessionRepository) Exit() {
	if r != nil {
		_ = r.client.Close()
	}
} 

package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Id uuid.UUID
	UserId uuid.UUID
	RefreshToken string
	IpAddress string
	UserAgent string
	CreatedAt time.Time
	ExpiresIn time.Time
}

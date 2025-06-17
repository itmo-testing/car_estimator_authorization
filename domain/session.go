package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserId    uuid.UUID `json:"userId"`
	Email 	  string `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	Source
}

type Source struct {
	IpAddress string `json:"addr"`
	UserAgent string `json:"ua"`
}
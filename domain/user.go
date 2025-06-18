package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserPublic
	Password string
	PasswordHash []byte
}

type UserPublic struct {
	Id uuid.UUID
	FullName string
	Email string
	Phone string
	BirthDate time.Time
	RegisterDate time.Time
}

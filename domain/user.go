package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id uuid.UUID
	FullName string
	Email string
	Phone string
	Password string
	BirthDate time.Time
	RegisterDate time.Time
}

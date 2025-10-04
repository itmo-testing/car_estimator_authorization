package unit

import (
	"io"
	"log/slog"
	"strings"

	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
	"golang.org/x/crypto/bcrypt"
)

func CreateTestUser(email, password string) *domain.User {
    passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    return &domain.User{
        UserPublic: domain.UserPublic{
            FullName: "Test User",
            Email:    email,
        },
        Password: password,
        PasswordHash: passwordHash,
    }
}

func IsValidJWTStructure(tokenString string) bool {
    parts := strings.Split(tokenString, ".")
    return len(parts) == 3
}

func NullLogger() *slog.Logger {
    return slog.New(slog.NewTextHandler(io.Discard, nil))
}

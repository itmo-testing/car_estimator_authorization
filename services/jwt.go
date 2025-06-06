package services

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
)

var secret = []byte(os.Getenv("SECRET_KEY"))

func CreateJWT(user *domain.User) (string, error) {
	payload := jwt.MapClaims{
        "sub":  user.Id,
		"email": user.Email,
		"iss": time.Now().Unix(),
        "exp":  time.Now().Add(time.Hour).Unix(),
    }

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	tokenString, err := token.SignedString(secret)

    if err != nil {
        return "", err
    }

	return tokenString, nil
}

func VerifyJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/joho/godotenv"
	pb "github.com/nikita-itmo-gh-acc/car_estimator_api_contracts/gen/profile_v1"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
)


type TestCase struct {
	name string
	user domain.User
	code codes.Code
	wantErr bool
}

func TestRegister(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("warning: can't find .env file")
	}

	tests := []TestCase{
		{
			name: "register new user",
			user: domain.User{
				FullName: "Ananiev Nikita",
				Email: "nikita-ananiev@mail.ru",
				Password: "qwertty",
				BirthDate: time.Date(2004, time.June, 24, 0, 0, 0, 0, time.Local),
			},
			code: codes.OK,
			wantErr: false,
		},
		{
			name: "missing email register",
			user: domain.User{
				FullName: "Shalunov Andrew",
				Password: "123465",
				BirthDate: time.Date(2003, time.December, 19, 0, 0, 0, 0, time.Local),
			},
			code: codes.InvalidArgument,
			wantErr: true,
		},
		{
			name: "duplicate email register",
			user: domain.User{
				FullName: "Ospelnikov Alex",
				Email: "nikita-ananiev@mail.ru",
				Password: "1qazxsw2",
				BirthDate: time.Date(2004, time.September, 17, 0, 0, 0, 0, time.Local),
			},
			code: codes.AlreadyExists,
			wantErr: true,
		},
	}

	ctx, client := NewClient(t, os.Getenv("APP_ADDR"))

	for idx, tt := range tests {
		fmt.Printf("[%d] name: %s\n", idx, tt.name)

		resp, err := client.Register(ctx, &pb.RegisterRequest{
			Fullname: tt.user.FullName,
			Email: tt.user.Email,
			Password: tt.user.Password,
			Birthdate: tt.user.BirthDate.Unix(),
		})

		if (err != nil) != tt.wantErr {
			t.Errorf("unexpected error: %v", err)
			continue
		}

		if !tt.wantErr && resp.GetUserId() == nil {
			t.Errorf("empty response")
			continue
		}

		code := status.Code(err)

		if code != tt.code {
			t.Errorf("unexpected status: want %v, have %v", tt.code, code)
		}
	}
}

package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/nikita-itmo-gh-acc/car_estimator_api_contracts/gen/profile_v1"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
)

func TestRegister(t *testing.T) {
	tests := []TestCase{
		{
			name: "register new user",
			args: domain.User{
				UserPublic: domain.UserPublic{
					FullName: "Ananiev Nikita",
					Email: "nikita-ananiev@mail.ru",
					Phone: "79111111111",
					BirthDate: time.Date(2004, time.June, 24, 0, 0, 0, 0, time.Local),
				},
				Password: "qwertty",
			},
			code: codes.OK,
			wantErr: false,
		},
		{
			name: "missing email and phone register",
			args: domain.User{
				UserPublic: domain.UserPublic{
					FullName: "Shalunov Andrew",
					BirthDate: time.Date(2003, time.December, 19, 0, 0, 0, 0, time.Local),
				},
				Password: "123465",
			},
			code: codes.InvalidArgument,
			wantErr: true,
		},
		{
			name: "duplicate email register",
			args: domain.User{
				UserPublic: domain.UserPublic{
					FullName: "Ospelnikov Alex",
					Email: "nikita-ananiev@mail.ru",
					Phone: "79999999999",
					BirthDate: time.Date(2004, time.September, 17, 0, 0, 0, 0, time.Local),
				},
				Password: "1qazxsw2",
			},
			code: codes.AlreadyExists,
			wantErr: true,
		},
	}

	defer CleanUpTestStorages(t, TestPGConn(t), nil)

	ctx, client := NewClient(t, os.Getenv("TEST_APP_ADDR"))

	for idx, tt := range tests {
		fmt.Printf("[%d] name: %s\n", idx, tt.name)
		user, ok := tt.args.(domain.User)
		if !ok {
			t.Errorf("unexpected args for the test")
			continue
		}

		resp, err := client.Register(ctx, &pb.RegisterRequest{
			Fullname: user.FullName,
			Email: user.Email,
			Phone: user.Phone,
			Password: user.Password,
			Birthdate: user.BirthDate.Unix(),
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

		fmt.Println("PASSED!")
	}
}

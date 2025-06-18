package tests

import (
	"fmt"
	"os"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/joho/godotenv"
	pb "github.com/nikita-itmo-gh-acc/car_estimator_api_contracts/gen/profile_v1"
)

type Credentials struct {
    email    string
    password string
}

func TestLoginAndRefresh(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("warning: can't find .env file")
	}

	tests := []TestCase{
		{
			name: "common login operation",
			args: Credentials{
				email: "nikita-ananiev@mail.ru",
				password: "qwertty",
			},
			code: codes.OK,
			wantErr: false,
		},
		{
			name: "login attempt with unregistered email",
			args: Credentials{
				email: "example@mail.ru",
				password: "12345",
			},
			code: codes.InvalidArgument,
			wantErr: true,
		},
		{
			name: "login attempt with wrong password",
			args: Credentials{
				email: "nikita-ananiev@mail.ru",
				password: "qwerty",
			},
			code: codes.InvalidArgument,
			wantErr: true,
		},
	}

	ctx, client := NewClient(t, os.Getenv("APP_ADDR"))
	mockAddr, mockUserAgent := "localhost:9999", "Chrome/137.0.0.0"

	for idx, tt := range tests {
		fmt.Printf("[%d] name: %s\n", idx, tt.name)

		credentials, ok := tt.args.(Credentials)
		if !ok {
			t.Errorf("unexpected args for the test")
			continue
		}

		resp, err := client.Login(ctx, &pb.LoginRequest{
			Email: credentials.email,
			Password: credentials.password,
			Source: &pb.SourceData{
				Ip: mockAddr,
				UserAgent: mockUserAgent,
			},
		})

		if (err != nil) != tt.wantErr {
			t.Errorf("unexpected error: %v", err)
			continue
		}

		if !tt.wantErr {
			if resp.GetAccessToken() == "" || resp.GetRefreshToken() == "" {
				t.Errorf("response has empty fields")
				continue
			}

			md := metadata.Pairs(
				"refreshToken", resp.GetRefreshToken(),
			)

    		ctx := metadata.NewOutgoingContext(ctx, md)
			_, err := client.Refresh(ctx, &pb.SourceData{
				Ip: mockAddr,
				UserAgent: mockUserAgent,
			})

			if err != nil {
				t.Errorf("unexpected error while do refresh op: %v", err)
			}

			continue
		}

		code := status.Code(err)

		if code != tt.code {
			t.Errorf("unexpected status: want %v, have %v", tt.code, code)
		}
	}
}

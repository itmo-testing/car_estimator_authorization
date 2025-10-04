package unit

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/database"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/services"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/tests/unit/mocks"

	mock "github.com/stretchr/testify/mock"
)


type Credentials struct {
    email    string
    password string
}

func TestUserLogin(t *testing.T) {
	// arrange
	var (
		validUser = CreateTestUser("test@test.ru", "123")

		testSource = domain.Source{
			IpAddress: "127.0.0.1:8000",
			UserAgent: "Chrome/137.0.0.0",
		}

		testRefreshToken = "Jk5p531mnd0k0OpJ9iZ9q9yZkq5Pub7o3QOWkIqasT8GcLK8GkgLHZQjEUbcvUKE"
		testUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	)

	validUser.Id = testUUID
	
	type Args struct {
		creds Credentials
		ctx context.Context
	}

	tests := []TestCase{
		{
			name: "Successful login",
			args: Args{
				creds: Credentials{
					email: "test@test.ru",
					password: "123",
				},
				ctx: context.Background(),
			},
			setupMocks: func(md *MockDependencies) {
				md.userProvider.
					On("Get", mock.Anything, "test@test.ru").
					Return(validUser, nil)
				
				md.sessionProvider.
					On("GetUserSessions", mock.Anything, testUUID).
					Return([]*domain.Session{}, nil)
				
				md.sessionSaver.
					On("Save", mock.Anything, mock.Anything, mock.Anything).
					Return(testRefreshToken, nil)
			},
			wantErr: false,
		},
		{
			name: "Failed login - no such user",
			args: Args{
				creds: Credentials{
					email: "absent@user.com",
					password: "123",
				},
				ctx: context.Background(),
			},
			setupMocks: func(md *MockDependencies) {
				md.userProvider.
					On("Get", mock.Anything, "absent@user.com").
					Return(nil, database.ErrUserNotFound)
			},
			wantErr: true,
		},
		{
			name: "Failed login - wrong password",
			args: Args{
				creds: Credentials{
					email: "test@test.ru",
					password: "111",
				},
				ctx: context.Background(),
			},
			setupMocks: func(md *MockDependencies) {
				md.userProvider.
					On("Get", mock.Anything, "test@test.ru").
					Return(validUser, nil)
			},
			wantErr: true,
		},
	}

	fmt.Println("========== Run login unit test ==========")

	for i, tt := range tests {
		fmt.Printf("[test #%d - %s]\n", i, tt.name)

		args, ok := tt.args.(Args)
		if !ok {
			t.Errorf("unexpected args for the test")
			continue
		}

		var (
			userProvider = mocks.NewIUserProvider(t)
			sessionProvider = mocks.NewISessionProvider(t)
			sessionSaver = mocks.NewISessionSaver(t)
			sessionRemover = mocks.NewISessionRemover(t)
		)
		
		md := &MockDependencies{
			userProvider: userProvider,	
			sessionProvider: sessionProvider,
			sessionSaver: sessionSaver,
			sessionRemover: sessionRemover,
		}

		tt.setupMocks(md)

		authService := services.NewAuthService(
			userProvider, sessionProvider, sessionSaver, sessionRemover, NullLogger(),
		)

		// act
		tokenPair, _, err := authService.Login(
			args.ctx, 
			args.creds.email,
			args.creds.password,
			testSource,
		)

		// assert
		if ((err != nil) != tt.wantErr) {
			t.Errorf("unexpected error: %v", err)
		}

		if (!tt.wantErr) {
			if (tokenPair.Access == "" || !IsValidJWTStructure(tokenPair.Access)) {
				t.Errorf("invalid access token: [%s]", tokenPair.Access)
			}

			if (tokenPair.Refresh == "") {
				t.Errorf("invalid refresh token: [%s]", tokenPair.Access)
			}
		}

		fmt.Println("PASSED!")
	}
}

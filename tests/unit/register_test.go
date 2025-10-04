package unit

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/database"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/services"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/tests/unit/mocks"

	mock "github.com/stretchr/testify/mock"
)

func TestUserRegister(t *testing.T) {
	type Args struct {
		ctx context.Context
		user domain.User
	}

	// arrange
	var (
		validUserCredentials = domain.User{
			UserPublic: domain.UserPublic{
				FullName: "Test User",
				Email: "test@test.ru",
				Phone: "+79999999999",
				BirthDate: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.Local),
			},
			Password: "qwerty",
		}

		invalidUserCredentials = validUserCredentials
		emptyId = uuid.UUID{}
	)
	
	invalidUserCredentials.Email = ""

	tests := []TestCase{
		{
			name: "Succsessful register",
			args: Args{
				ctx: context.Background(),
				user: validUserCredentials,
			},
			setupMocks: func(md *MockDependencies) {
				md.userSaver.
					On("Save", mock.Anything, mock.AnythingOfType("domain.User")).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "Register attempt again",
			args: Args{
				ctx: context.Background(),
				user: validUserCredentials,
			},
			setupMocks: func(md *MockDependencies) {
				md.userSaver.
					On("Save", mock.Anything, mock.AnythingOfType("domain.User")).
					Return(database.ErrUserAlreadyExists)
			},
			wantErr: true,
		},
		{
			name: "Invalid email",
			args: Args{
				ctx: context.Background(),
				user: invalidUserCredentials,
			},
			setupMocks: func(md *MockDependencies) {
				md.userSaver.
					On("Save", mock.Anything, mock.AnythingOfType("domain.User")).
					Return(errors.New("Invalid email value"))
			},
			wantErr: true,
		},
	}

	fmt.Println("========== Run register unit test ==========")

	for i, tt := range tests {
		fmt.Printf("[test #%d - %s]\n", i, tt.name)

		args, ok := tt.args.(Args)
		if !ok {
			t.Errorf("unexpected args for the test")
			continue
		}

		var (
			userSaver = mocks.NewIUserSaver(t)
			userRemover = mocks.NewIUserRemover(t)
			sessionProvider = mocks.NewISessionProvider(t)
			sessionRemover = mocks.NewISessionRemover(t)
		)
		
		md := &MockDependencies{
			userSaver: userSaver,
		}

		tt.setupMocks(md)

		registerService := services.NewRegistrarService(
			userSaver, userRemover, sessionRemover, sessionProvider, NullLogger(),
		)

		// act
		userId, err := registerService.Register(
			args.ctx,
			args.user,
		)

		// assert
		if ((err != nil) != tt.wantErr) {
			t.Errorf("unexpected error: %v", err)
		}

		if (!tt.wantErr) {
			if (userId == emptyId) {
				t.Errorf("register returned empty uuid")
			}
		}

		fmt.Println("PASSED!")
	}
}

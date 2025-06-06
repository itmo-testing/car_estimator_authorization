package grpc_server

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	pb "github.com/nikita-itmo-gh-acc/car_estimator_api_contracts/gen"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/services"
)

type ServerAPI struct {
	Auth IAuthService
	pb.UnimplementedAuthServiceServer
}

type IAuthService interface {
	Login(ctx context.Context, email string, password string) (token string, err error)
	Register(ctx context.Context, user domain.User) (userId uuid.UUID, err error)
	// Refresh(ctx context.Context, )
}

func RegisterServer(srv *grpc.Server, auth IAuthService) {
	pb.RegisterAuthServiceServer(srv, &ServerAPI{ Auth: auth })
}

func (s *ServerAPI) Login(ctx context.Context, in *pb.LoginRequest) (t *pb.TokenResponse, err error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	token, err := s.Auth.Login(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "wrong email or password")
		}

		return nil, status.Error(codes.Internal, "login failed")
	}

	return &pb.TokenResponse{AccessToken: token}, nil
}

func (s *ServerAPI) Register(ctx context.Context, in *pb.RegisterRequest) (t *pb.RegisterResponse, err error) {
	msg := in.ProtoReflect()
	fields := msg.Descriptor().Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if !msg.Has(field) {
			return nil, status.Error(codes.InvalidArgument, field.TextName() + " is required")
		}
	}

	newUser := domain.User{
		FullName: in.Fullname,
		Email: in.Email,
		Password: in.Password,
		BirthDate: time.Unix(in.Birthdate, 0),
	}

	userId, err := s.Auth.Register(ctx, newUser)
	if err != nil {
		return nil, status.Error(codes.Internal, "registration failed")
	}

	uuid := &pb.UUID{Value: userId.String()}
	return &pb.RegisterResponse{UserId: uuid}, nil
}

func (s *ServerAPI) Refresh(ctx context.Context, in *pb.RefreshRequest) (t *pb.TokenResponse, err error) {
	return &pb.TokenResponse{}, nil // пока что заглушка
}

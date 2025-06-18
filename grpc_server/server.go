package grpc_server

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/google/uuid"
	pb "github.com/nikita-itmo-gh-acc/car_estimator_api_contracts/gen/profile_v1"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/database"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/services"
)

type ServerAPI struct {
	Auth IAuthService
	Registrar IRegistrarSesvice
	pb.UnimplementedProfileServiceServer
}

type IAuthService interface {
	Login(ctx context.Context, email, password string, source domain.Source) (*domain.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string, source domain.Source) (*domain.TokenPair, error)
}

type IRegistrarSesvice interface {
	Register(ctx context.Context, user domain.User) (userId uuid.UUID, err error)
	Unregister(ctx context.Context, refreshToken string) error
}

func RegisterServer(srv *grpc.Server, auth IAuthService, reg IRegistrarSesvice) {
	pb.RegisterProfileServiceServer(srv, &ServerAPI{ Auth: auth, Registrar: reg })
}

func GetRefreshToken(ctx context.Context) (string, error) {
	val := ctx.Value("refreshToken")
	if val == nil {
		return "", status.Error(codes.Unauthenticated, "can't find refresh token")
	}

	rt, ok := val.(string)
	if !ok {
		return "", status.Error(codes.InvalidArgument, "refresh token must be string")
	}

	return rt, nil
}

func (s *ServerAPI) Login(ctx context.Context, in *pb.LoginRequest) (t *pb.TokenResponse, err error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	data := in.GetSource()

	tokens, err := s.Auth.Login(ctx, in.GetEmail(), in.GetPassword(), domain.Source{IpAddress: data.Ip, UserAgent: data.UserAgent})
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "wrong email or password")
		}

		return nil, status.Error(codes.Internal, "login failed")
	}

	return &pb.TokenResponse{
		AccessToken: tokens.Access, 
		RefreshToken: tokens.Refresh,
	}, nil
}

func (s *ServerAPI) Logout(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	refreshToken, err := GetRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	if err = s.Auth.Logout(ctx, refreshToken); err != nil {
		if errors.Is(err, database.ErrSessionNotFound) {
			return nil, status.Error(codes.Unauthenticated, "user session not found")
		}
		return nil, status.Error(codes.Internal, "logout failed")
	}

	return &emptypb.Empty{}, nil
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

	userId, err := s.Registrar.Register(ctx, newUser)
	if err != nil {
		if errors.Is(err, database.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "registration failed")
	}

	uuid := &pb.UUID{Value: userId.String()}
	return &pb.RegisterResponse{UserId: uuid}, nil
}

func (s *ServerAPI) Unregister(ctx context.Context, in *pb.UnregiserRequest) (*emptypb.Empty, error) {
	refreshToken, err := GetRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	if err = s.Registrar.Unregister(ctx, refreshToken); err != nil {
		if errors.Is(err, database.ErrSessionNotFound) {
			return nil, status.Error(codes.Unauthenticated, "user session not found")
		}

		return nil, status.Error(codes.Internal, "user unregister failed")
	}

	return &emptypb.Empty{}, nil
}

func (s *ServerAPI) Refresh(ctx context.Context, in *pb.SourceData) (t *pb.TokenResponse, err error) {
	refreshToken, err := GetRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	source := domain.Source{
		IpAddress: in.GetIp(),
		UserAgent: in.GetUserAgent(),
	}

	tokens, err := s.Auth.Refresh(ctx, refreshToken, source)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrSessionNotFound):
			return nil,	status.Error(codes.Unauthenticated, "user session not found")
		case errors.Is(err, services.ErrSourceChanged):
			return nil, status.Error(codes.PermissionDenied, "attempt to enter from unknown device")
		default:
			return nil, status.Error(codes.Internal, "tokens refresh failed")
		}
	}

	return &pb.TokenResponse{
		AccessToken: tokens.Access, 
		RefreshToken: tokens.Refresh,
	}, nil
}

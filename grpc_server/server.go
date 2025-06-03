package grpc_server

import (
	"context"
	"github.com/google/uuid"

    "google.golang.org/grpc"
	pb "github.com/nikita-itmo-gh-acc/car_estimator_authorization/gen"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
)

type ServerAPI struct {
	Auth IAuthService
	pb.UnimplementedAuthServiceServer
}

type IAuthService interface {
	Login(ctx context.Context, email string, password string) (token string, err error)
	Register(ctx context.Context, user domain.User) (userId uuid.UUID, err error)
	Refresh(ctx context.Context, )
}

func RegisterServer(srv *grpc.Server, auth IAuthService) {
	pb.RegisterAuthServiceServer(srv, &ServerAPI{ Auth: auth })
}

func (s *ServerAPI) Login(ctx context.Context, in *pb.LoginRequest) (t *pb.TokenResponse, err error) {
	
}

func (s *ServerAPI) Register(ctx context.Context, in *pb.RefreshRequest) (t *pb.RegisterResponse, err error) {

}

func ()
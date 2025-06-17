package app

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/database"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/grpc_server"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/interceptors"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type repository interface {
	Exit()
}

type App struct {
	logger *slog.Logger
	users repository
	sessions repository
	gRPCserver *grpc.Server
	port int
}

func New(logger *slog.Logger, userStorageConfig, sessionStorageConfig *database.Config, port int) (*App, error) {
	userRepository, err := database.NewUserRepository(userStorageConfig)
	if err != nil {
		return nil, err
	}

	sessionRepository, err := database.NewSessionRepository(sessionStorageConfig)
	if err != nil {
		return nil, err
	}

	authService := services.NewAuthService(
		userRepository, 
		sessionRepository,
		sessionRepository,
		sessionRepository,
		logger,
	)

	registrarService := services.NewRegistrarService(
		userRepository,
		userRepository,
		sessionRepository,
		sessionRepository,
		logger,
	)

	recoveryOpts := []recovery.Option{
        recovery.WithRecoveryHandler(func(p interface{}) (err error) {
            logger.Error("Recovered from panic", slog.Any("panic", p))

            return status.Errorf(codes.Internal, "internal error")
        }),
    }

	privateHandlers := map[string]struct{} {
		"/profile.ProfileService/Unregister": {},
		"/profile.ProfileService/Logout": {},
		"/profile.ProfileService/Refresh": {},
	}

	chain := grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		interceptors.RefreshTokenInterceptor(privateHandlers),
		interceptors.SlogUnaryServerInterceptor(logger),
	)

	server := grpc.NewServer(chain)
	grpc_server.RegisterServer(server, authService, registrarService)

	return &App{
		logger: logger,
		users: userRepository,
		sessions: sessionRepository,
		gRPCserver: server,
		port: port,
	}, nil
}

func (app *App) Run() error {
	app.logger.Info("grpc server run attempt...")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", app.port))
	if err != nil {
		return fmt.Errorf("run failed: %w", err)
	}

	app.logger.Info("grpc server started!")

	if err := app.gRPCserver.Serve(l); err != nil {
        return fmt.Errorf("run failed: %w", err)
    }

	return nil
}

func (app *App) Stop() {
	if app.users != nil {
		app.users.Exit()
	}

	if app.sessions != nil {
		app.sessions.Exit()
	}

	app.gRPCserver.GracefulStop()
}
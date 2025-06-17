package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)


func SlogUnaryServerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
    ) (resp interface{}, err error) {
        start := time.Now()
        md, _ := metadata.FromIncomingContext(ctx)
        
        logger.Info("gRPC request started",
            "method", info.FullMethod,
            "metadata", md,
        )

        resp, err = handler(ctx, req)

        duration := time.Since(start)
        code := status.Code(err)

        logger.LogAttrs(ctx, slog.LevelInfo, "gRPC request completed",
            slog.String("method", info.FullMethod),
            slog.String("duration", duration.String()),
            slog.String("status", code.String()),
            slog.Any("error", err),
        )

        return resp, err
    }
}

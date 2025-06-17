package interceptors

import (
	"context"

	"google.golang.org/grpc"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/metadata"
	// "google.golang.org/grpc/status"
)

func AuthInterceptor(privateHandlers map[string]struct{}) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		return nil, nil
	}
}
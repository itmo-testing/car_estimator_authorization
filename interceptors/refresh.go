package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func RefreshTokenInterceptor(privateHandlers map[string]struct{}) grpc.UnaryServerInterceptor { 
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := privateHandlers[info.FullMethod]; !ok {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is missing")
		}

		values := md.Get("refreshToken")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "refresh token missing")
		}

		ctx = context.WithValue(ctx, "refreshToken", values[0])

		return handler(ctx, req)
	}
}

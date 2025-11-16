package integration

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/nikita-itmo-gh-acc/car_estimator_api_contracts/gen/profile_v1"
)

var (
	defaultTestServerAddr = "test_container:4445"
)

func NewClient(t *testing.T, addr string) (context.Context, pb.ProfileServiceClient) {
	t.Helper()

	if addr == "" {
		addr = defaultTestServerAddr
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)

	t.Cleanup(func() {
        t.Helper()
        cancel()
    })

	cc, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	client := pb.NewProfileServiceClient(cc)

	return ctx, client
}

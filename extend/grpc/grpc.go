package cherryGRPC

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"time"
)

func GetIP(ctx context.Context) string {
	p, found := peer.FromContext(ctx)
	if found == false {
		return ""
	}
	return p.Addr.String()
}

func BuildRpcClientConn(address string, timeout time.Duration, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if timeout.Seconds() < 1 {
		timeout = 2 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	dialConn, err := grpc.DialContext(
		ctx,
		address,
		opts...,
	)
	cancel()

	if err != nil {
		return nil, err
	}

	return dialConn, nil
}

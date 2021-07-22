package cherryGRPC

import (
	"context"
	"google.golang.org/grpc/peer"
)

func GetIP(ctx context.Context) string {
	p, found := peer.FromContext(ctx)
	if found == false {
		return ""
	}
	return p.Addr.String()
}

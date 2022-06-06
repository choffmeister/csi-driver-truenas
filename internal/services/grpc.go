package services

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/choffmeister/csi-driver-truenas/internal/utils"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

func CreateCSIListener() (net.Listener, error) {
	endpoint := os.Getenv("CSI_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("you need to specify an endpoint via the CSI_ENDPOINT env var")
	}
	if !strings.HasPrefix(endpoint, "unix://") {
		return nil, fmt.Errorf("endpoint must start with unix://")
	}
	utils.Debug.Printf("Listening on %s\n", endpoint)
	socketFile := strings.TrimPrefix(endpoint, "unix://")

	if err := os.Remove(socketFile); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove socket file at %s: %s", socketFile, err)
	}

	return net.Listen("unix", socketFile)
}

func CreateGRPCServer() *grpc.Server {
	requestLogger := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		isProbe := info.FullMethod == "/csi.v1.Identity/Probe"

		resp, err := handler(ctx, req)
		if err != nil {
			utils.Error.Printf("Handling request %v %v failed: %v\n", info.FullMethod, req, err)
		} else if !isProbe {
			utils.Debug.Printf("Handled request %v %v", info.FullMethod, req)
		}
		return resp, err
	}

	return grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				requestLogger,
			),
		),
	)
}

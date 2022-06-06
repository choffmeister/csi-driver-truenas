package cmd

import (
	"fmt"
	"os"

	"github.com/choffmeister/csi-driver-truenas/internal/services"
	proto "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/spf13/cobra"
)

var (
	nodeCmd = &cobra.Command{
		Use: "node",
		RunE: func(cmd *cobra.Command, args []string) error {
			listener, err := services.CreateCSIListener()
			if err != nil {
				return err
			}
			grpcServer := services.CreateGRPCServer()

			identityService := services.NewIdentityService()
			proto.RegisterIdentityServer(grpcServer, identityService)

			nodeId := os.Getenv("KUBE_NODE_NAME")
			if nodeId == "" {
				return fmt.Errorf("you need to specify the node nama via the KUBE_NODE_NAME env var")
			}

			nodeService := services.NewNodeService(nodeId)
			proto.RegisterNodeServer(grpcServer, nodeService)

			identityService.SetReady(true)
			if err := grpcServer.Serve(listener); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
}

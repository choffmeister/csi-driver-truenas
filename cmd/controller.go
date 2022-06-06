package cmd

import (
	"github.com/choffmeister/csi-driver-truenas/internal/services"
	proto "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/spf13/cobra"
)

var (
	controllerCmd = &cobra.Command{
		Use: "controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			listener, err := services.CreateCSIListener()
			if err != nil {
				return err
			}
			grpcServer := services.CreateGRPCServer()

			identityService := services.NewIdentityService()
			proto.RegisterIdentityServer(grpcServer, identityService)

			controllerService := services.NewControllerService()
			proto.RegisterControllerServer(grpcServer, controllerService)

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

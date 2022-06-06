package cmd

import (
	"io/ioutil"
	"log"

	"github.com/choffmeister/csi-driver-truenas/internal/utils"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	rootCmd = &cobra.Command{
		Use: "csi-driver-truenas",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !verbose {
				utils.Debug = log.New(ioutil.Discard, "", log.LstdFlags)
			}
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(controllerCmd)
	rootCmd.AddCommand(nodeCmd)
}

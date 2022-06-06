package main

import (
	"os"

	"github.com/choffmeister/csi-driver-truenas/cmd"
	"github.com/choffmeister/csi-driver-truenas/internal/utils"
)

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	cmd.Version = cmd.FullVersion{Version: version, Commit: commit, Date: date, BuiltBy: builtBy}
	if err := cmd.Execute(); err != nil {
		utils.Error.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

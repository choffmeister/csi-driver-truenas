package e2e

import (
	"os"
	"testing"

	"github.com/choffmeister/csi-driver-truenas/test"
)

func TestMain(m *testing.M) {
	env := test.LoadTestEnv()
	name := "csi-driver-test"
	if os.Getenv("CLUSTER_START") == "true" {
		if err := env.StartTalosCluster(env.TalosIP, name); err != nil {
			panic(err)
		}
	}

	rc := m.Run()

	if os.Getenv("CLUSTER_STOP") == "true" {
		if err := env.StopTalosCluster(); err != nil {
			panic(err)
		}
	}

	os.Exit(rc)
}

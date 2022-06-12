package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/choffmeister/csi-driver-truenas/test"
)

var (
	kubectlBin = test.KubectlBin
	kubeConfig = test.KubeConfig
)

func TestMain(m *testing.M) {
	if err := test.PrepareAllBinaries(); err != nil {
		panic(err)
	}

	env := test.LoadTestEnv("../../test.env")
	name := "csi-driver-test"
	if os.Getenv("CLUSTER_START") == "true" {
		if err := test.StartTalosCluster(env.TalosIP, name); err != nil {
			panic(err)
		}
	}

	rc := m.Run()

	if os.Getenv("CLUSTER_STOP") == "true" {
		if err := test.StopTalosCluster(); err != nil {
			panic(err)
		}
	}

	os.Exit(rc)
}

func execKubectl(args []string, stdin string) (string, int, error) {
	fullArgs := []string{"--kubeconfig", kubeConfig}
	fullArgs = append(fullArgs, args...)
	opts := test.ExecOpts{
		Name:       kubectlBin,
		Args:       fullArgs,
		Input:      stdin,
		Output:     os.Stdout,
		Retries:    300 / 5,
		RetryDelay: 5 * time.Second,
	}
	return test.ExecWithOpts(opts)
}

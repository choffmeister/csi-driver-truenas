package e2e

import (
	"fmt"
	"log"
	"os"
	"path"
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
		log.Panic(err)
	}

	env := test.LoadTestEnv()
	name := "csi-driver-test"
	if os.Getenv("CLUSTER_START") == "true" {
		if err := test.StartTalosCluster(env.TalosIP, name); err != nil {
			log.Panic(err)
		}
	}

	rc := m.Run()

	if os.Getenv("CLUSTER_STOP") == "true" {
		if err := test.StopTalosCluster(); err != nil {
			log.Panic(err)
		}
	}

	os.Exit(rc)
}

func TestSimple(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}
	if _, _, err := test.ExecRetryWithOutput(kubectlBin, "--kubeconfig", kubeConfig, "apply", "-k", path.Join(dir, "manifests")); err != nil {
		t.Error(err)
		return
	}

	env := test.LoadTestEnv()
	storageClassSecretYaml := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
    name: csi-driver-truenas-volumes
    namespace: kube-system
stringData:
    truenas-url: "%s"
    truenas-api-key: "%s"
    truenas-tls-skip-verify: "true"
    truenas-parent-dataset: "%s"
    iscsi-base-iqn: "%s"
    iscsi-portal-ip: "%s"
    iscsi-portal-port: "%s"
    iscsi-portal-id: "%s"
    iscsi-initiator-id: "%s"
`,
		env.TruenasUrl,
		env.TruenasApiKey,
		env.TruenasParentDataset,
		env.ISCSIBaseIQN,
		env.ISCSIPortalIP,
		env.ISCSIPortalPort,
		env.ISCSIPortalID,
		env.ISCSIInitiatorID,
	)
	cmdOpts := test.ExecCommandOpts{
		Name:       kubectlBin,
		Args:       []string{"--kubeconfig", kubeConfig, "apply", "-f", "-"},
		Output:     os.Stdout,
		Retries:    300 / 5,
		RetryDelay: 5 * time.Second,
		Stdin:      storageClassSecretYaml,
	}
	if _, _, err := test.ExecCommandWithOpts(cmdOpts); err != nil {
		t.Error(err)
		return
	}
}

func TestOfficialTestsuite(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(t)
		return
	}
	testDriverYaml := path.Join(cwd, "test-driver.yaml")

	// TODO the only test matched by this fails
	// t.Run("serial tests", func(t *testing.T) {
	// 	_, _, err := test.ExecWithOutput(test.E2EGinkgoBin, "-v", "-focus=External.Storage.*(\\[Feature:|\\[Serial\\])", "-skip=\\[Disruptive\\]", test.E2ETestBin, "--", "-storage.testdriver="+testDriverYaml)
	// 	if err != nil {
	// 		t.Error(err)
	// 	}
	// })

	t.Run("parallel tests", func(t *testing.T) {
		_, _, err := test.ExecWithOutput(test.E2EGinkgoBin, "-v", "-nodes=4", "-focus=External.Storage", "-skip=\\[Feature:|\\[Disruptive\\]|\\[Serial\\]", test.E2ETestBin, "--", "-storage.testdriver="+testDriverYaml)
		if err != nil {
			t.Error(err)
		}
	})
}

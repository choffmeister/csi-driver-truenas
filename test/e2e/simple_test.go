package e2e

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	_ "embed"

	"github.com/choffmeister/csi-driver-truenas/test"
	"github.com/choffmeister/csi-driver-truenas/test/e2e/manifests"
)

func TestSimple(t *testing.T) {
	env := test.LoadTestEnv()
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}

	// installing csi-driver-truenas
	if _, _, err := execKubectl(&env, []string{"apply", "-k", path.Join(dir, "manifests")}, ""); err != nil {
		t.Error(err)
		return
	}
	if _, _, err := execKubectl(&env, []string{"apply", "-f", "-"}, test.RenderTemplateFromEnv(env, manifests.SecretTemplate)); err != nil {
		t.Error(err)
		return
	}
	if _, _, err := execKubectl(&env, []string{"apply", "-f", "-"}, test.RenderTemplateFromEnv(env, manifests.SecretCIFSTemplate)); err != nil {
		t.Error(err)
		return
	}

	// creating test consumer
	if _, _, err := execKubectl(&env, []string{"apply", "-f", "-"}, test.RenderTemplateFromEnv(env, manifests.ConsumerTemplate)); err != nil {
		t.Error(err)
		return
	}

	// waiting for test consumer to be running
	if err := test.Retry(func() error {
		if output, _, err := execKubectl(&env, []string{"get", "pod", "csi-driver-truenas-consumer-0", "-n", "csi-driver-truenas"}, ""); err != nil {
			return err
		} else if !strings.Contains(output, "Running") {
			return fmt.Errorf("pod csi-driver-truenas-consumer-0 is not yet running")
		}
		return nil
	}, 300, time.Second); err != nil {
		t.Error(err)
		return
	}

	// ensure test consumer to have volume mounted
	if output, _, err := execKubectl(&env, []string{"exec", "csi-driver-truenas-consumer-0", "-n", "csi-driver-truenas", "--", "ls", "/mnt/data"}, ""); err != nil {
		t.Error(err)
		return
	} else if !strings.Contains(output, "lost+found") {
		t.Error(fmt.Errorf("pod csi-driver-truenas-consumer-0 does not see the persistent volume correctly"))
		return
	}

	// capture output of controller and node
	if output, _, err := execKubectl(&env, []string{"logs", "-l", "app=csi-driver-truenas-csi-controller", "-n", "csi-driver-truenas", "-c", "csi-driver-truenas-csi-driver"}, ""); err != nil {
		t.Error(err)
		return
	} else {
		fmt.Printf("Controller logs\n%s\n", output)
	}
	if output, _, err := execKubectl(&env, []string{"logs", "-l", "app=csi-driver-truenas-csi-node", "-n", "csi-driver-truenas", "-c", "csi-driver-truenas-csi-driver"}, ""); err != nil {
		t.Error(err)
		return
	} else {
		fmt.Printf("Node logs\n%s\n", output)
	}

	// deleting test consumer and its persistent volume claim
	if _, _, err := execKubectl(&env, []string{"delete", "statefulset", "csi-driver-truenas-consumer", "-n", "csi-driver-truenas"}, ""); err != nil {
		t.Error(err)
		return
	}
	if _, _, err := execKubectl(&env, []string{"delete", "persistentvolumeclaim", "data-csi-driver-truenas-consumer-0", "-n", "csi-driver-truenas"}, ""); err != nil {
		t.Error(err)
		return
	}

	// waiting for test consumer persistent volume to be deleted
	if err := test.Retry(func() error {
		if output, _, err := execKubectl(&env, []string{"get", "persistentvolume"}, ""); err != nil {
			return err
		} else if strings.Contains(output, "data-csi-driver-truenas-consumer-0") {
			return fmt.Errorf("persistent volume data-csi-driver-truenas-consumer-0 is not yet deleted")
		}
		return nil
	}, 300, time.Second); err != nil {
		t.Error(err)
		return
	}

	// uninstalling csi-driver-truenas
	if _, _, err := execKubectl(&env, []string{"delete", "-f", "-"}, test.RenderTemplateFromEnv(env, manifests.SecretTemplate)); err != nil {
		t.Error(err)
		return
	}
	if _, _, err := execKubectl(&env, []string{"delete", "-k", path.Join(dir, "manifests")}, ""); err != nil {
		t.Error(err)
		return
	}
}

func execKubectl(env *test.TestEnv, args []string, stdin string) (string, int, error) {
	fullArgs := []string{"--kubeconfig", env.KubeConfig}
	fullArgs = append(fullArgs, args...)
	opts := test.ExecOpts{
		Name:       env.KubectlBin,
		Args:       fullArgs,
		Input:      stdin,
		Output:     os.Stdout,
		Retries:    300 / 5,
		RetryDelay: 5 * time.Second,
	}
	return test.ExecWithOpts(opts)
}

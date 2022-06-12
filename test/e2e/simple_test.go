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
	env := test.LoadTestEnv("../../test.env")
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}

	// installing csi-driver-truenas
	if _, _, err := execKubectl([]string{"apply", "-k", path.Join(dir, "manifests")}, ""); err != nil {
		t.Error(err)
		return
	}
	if _, _, err := execKubectl([]string{"apply", "-f", "-"}, test.RenderTemplateFromEnv(env, manifests.SecretTemplate)); err != nil {
		t.Error(err)
		return
	}

	// creating test consumer
	if _, _, err := execKubectl([]string{"apply", "-f", "-"}, test.RenderTemplateFromEnv(env, manifests.ConsumerTemplate)); err != nil {
		t.Error(err)
		return
	}

	// waiting for test consumer to be running
	if err := test.Retry(func() error {
		if output, _, err := execKubectl([]string{"get", "pod", "csi-driver-truenas-consumer-0", "-n", "csi-driver-truenas"}, ""); err != nil {
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
	if output, _, err := execKubectl([]string{"exec", "csi-driver-truenas-consumer-0", "-n", "csi-driver-truenas", "--", "ls", "/mnt/data"}, ""); err != nil {
		t.Error(err)
		return
	} else if !strings.Contains(output, "lost+found") {
		t.Error(fmt.Errorf("pod csi-driver-truenas-consumer-0 does not see the persistent volume correctly"))
		return
	}

	// deleting test consumer and its persistent volume claim
	if _, _, err := execKubectl([]string{"delete", "statefulset", "csi-driver-truenas-consumer", "-n", "csi-driver-truenas"}, ""); err != nil {
		t.Error(err)
		return
	}
	if _, _, err := execKubectl([]string{"delete", "persistentvolumeclaim", "data-csi-driver-truenas-consumer-0", "-n", "csi-driver-truenas"}, ""); err != nil {
		t.Error(err)
		return
	}

	// waiting for test consumer persistent volume to be deleted
	if err := test.Retry(func() error {
		if output, _, err := execKubectl([]string{"get", "persistentvolume"}, ""); err != nil {
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
	if _, _, err := execKubectl([]string{"delete", "-f", "-"}, test.RenderTemplateFromEnv(env, manifests.SecretTemplate)); err != nil {
		t.Error(err)
		return
	}
	if _, _, err := execKubectl([]string{"delete", "-k", path.Join(dir, "manifests")}, ""); err != nil {
		t.Error(err)
		return
	}
}

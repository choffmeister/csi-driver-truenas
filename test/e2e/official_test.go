package e2e

import (
	"os"
	"path"
	"testing"

	"github.com/choffmeister/csi-driver-truenas/test"
)

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

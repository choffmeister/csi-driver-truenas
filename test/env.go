package test

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path"
	"runtime"
	"text/template"

	"github.com/joho/godotenv"
)

var (
	kubernetesVersion = "1.24.1"
	talosVersion      = "1.0.5"
	e2eUrl            = fmt.Sprintf("https://dl.k8s.io/v%s/kubernetes-test-%s-%s.tar.gz", kubernetesVersion, runtime.GOOS, runtime.GOARCH)
	kubectlUrl        = fmt.Sprintf("https://dl.k8s.io/release/v%s/bin/%s/%s/kubectl", kubernetesVersion, runtime.GOOS, runtime.GOARCH)
	talosctlUrl       = fmt.Sprintf("https://github.com/siderolabs/talos/releases/download/v%s/talosctl-%s-%s", talosVersion, runtime.GOOS, runtime.GOARCH)
)

type TestEnv struct {
	BaseDir string
	TempDir string

	E2ETestBin   string
	E2EGinkgoBin string
	KubectlBin   string
	TalosctlBin  string

	ClusterDir  string
	TalosConfig string
	KubeConfig  string

	TalosIP              string
	TruenasUrl           string
	TruenasApiKey        string
	TruenasParentDataset string
	ISCSIBaseIQN         string
	ISCSIPortalIP        string
	ISCSIPortalPort      string
	ISCSIPortalID        string
	ISCSIInitiatorID     string
	CIFSIP               string
	CIFSShare            string
	CIFSUsername         string
	CIFSPassword         string
}

func LoadTestEnv() TestEnv {
	env := TestEnv{}

	if _, caller, _, ok := runtime.Caller(0); !ok {
		panic(fmt.Errorf("unable to detect caller"))
	} else {
		env.BaseDir = path.Join(path.Dir(caller), "..")
		env.TempDir = path.Join(env.BaseDir, ".tmp")
	}

	if err := godotenv.Load(path.Join(env.BaseDir, "test.env")); err != nil {
		panic(err)
	}

	env.E2ETestBin = path.Join(env.TempDir, fmt.Sprintf("e2e-test-%s.test", kubernetesVersion))
	env.E2EGinkgoBin = path.Join(env.TempDir, fmt.Sprintf("e2e-ginkgo-%s", kubernetesVersion))
	env.KubectlBin = path.Join(env.TempDir, fmt.Sprintf("kubectl-%s", kubernetesVersion))
	env.TalosctlBin = path.Join(env.TempDir, fmt.Sprintf("talosctl-%s", talosVersion))
	binaryPacks := map[string]BinariesUnpack{
		e2eUrl: &TarGzBinariesUnpack{
			Entries: map[string]string{
				"kubernetes/test/bin/e2e.test": env.E2ETestBin,
				"kubernetes/test/bin/ginkgo":   env.E2EGinkgoBin,
			},
		},
		kubectlUrl: &RawBinariesUnpack{
			Name: env.KubectlBin,
		},
		talosctlUrl: &RawBinariesUnpack{
			Name: env.TalosctlBin,
		},
	}

	for url, unpack := range binaryPacks {
		if err := PrepareBinaries(url, unpack); err != nil {
			panic(err)
		}
	}

	env.ClusterDir = path.Join(env.TempDir, "cluster")
	env.TalosConfig = path.Join(env.ClusterDir, "talosconfig")
	env.KubeConfig = path.Join(env.ClusterDir, "kubeconfig")

	env.TalosIP = os.Getenv("TALOS_IP")
	if env.TalosIP == "" {
		panic("env TALOS_IP is missing")
	}

	env.TruenasUrl = os.Getenv("TRUENAS_URL")
	if env.TruenasUrl == "" {
		panic("env TRUENAS_URL is missing")
	}

	env.TruenasApiKey = os.Getenv("TRUENAS_API_KEY")
	if env.TruenasApiKey == "" {
		panic("env TRUENAS_API_KEY is missing")
	}

	env.TruenasParentDataset = os.Getenv("TRUENAS_PARENT_DATASET")
	if env.TruenasParentDataset == "" {
		panic("env TRUENAS_PARENT_DATASET is missing")
	}

	env.ISCSIBaseIQN = os.Getenv("ISCSI_BASE_IQN")
	if env.ISCSIBaseIQN == "" {
		panic("env ISCSI_BASE_IQN is missing")
	}

	env.ISCSIPortalIP = os.Getenv("ISCSI_PORTAL_IP")
	if env.ISCSIPortalIP == "" {
		panic("env ISCSI_PORTAL_IP is missing")
	}

	env.ISCSIPortalPort = os.Getenv("ISCSI_PORTAL_PORT")
	if env.ISCSIPortalPort == "" {
		panic("env ISCSI_PORTAL_PORT is missing")
	}

	env.ISCSIPortalID = os.Getenv("ISCSI_PORTAL_ID")
	if env.ISCSIPortalID == "" {
		panic("env ISCSI_PORTAL_ID is missing")
	}

	env.ISCSIInitiatorID = os.Getenv("ISCSI_INITIATOR_ID")
	if env.ISCSIInitiatorID == "" {
		panic("env ISCSI_INITIATOR_ID is missing")
	}

	env.CIFSIP = os.Getenv("CIFS_IP")
	if env.CIFSIP == "" {
		panic("env CIFS_IP is missing")
	}

	env.CIFSShare = os.Getenv("CIFS_SHARE")
	if env.CIFSShare == "" {
		panic("env CIFS_SHARE is missing")
	}

	env.CIFSUsername = os.Getenv("CIFS_USERNAME")
	if env.CIFSUsername == "" {
		panic("env CIFS_USERNAME is missing")
	}

	env.CIFSPassword = os.Getenv("CIFS_PASSWORD")
	if env.CIFSPassword == "" {
		panic("env CIFS_PASSWORD is missing")
	}

	return env
}

func RenderTemplateFromEnv(env TestEnv, templateStr string) string {
	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		panic(err)
	}
	data := map[string]interface{}{
		"Env": env,
	}
	writer := bytes.Buffer{}
	if err := tmpl.Execute(&writer, data); err != nil {
		panic(err)
	}
	return writer.String()
}

//go:embed cluster-patch.yaml
var talosPatchYaml string

func (env *TestEnv) StartTalosCluster(ip string, name string) error {
	fmt.Printf("Starting talos cluster\n")

	if _, err := os.Stat(env.TalosConfig); err != nil && os.IsNotExist(err) {
		if _, _, err := Exec(env.TalosctlBin, "gen", "config", name, "https://"+ip+":6443", "--config-patch", talosPatchYaml, "--output-dir", env.ClusterDir); err != nil {
			return err
		}
	}
	if _, _, err := ExecRetry(env.TalosctlBin, "apply-config", "--insecure", "--nodes", ip, "--file", path.Join(env.ClusterDir, "controlplane.yaml")); err != nil {
		return err
	}
	if _, _, err := ExecRetry(env.TalosctlBin, "--talosconfig", env.TalosConfig, "config", "endpoint", ip); err != nil {
		return err
	}
	if _, _, err := ExecRetry(env.TalosctlBin, "--talosconfig", env.TalosConfig, "config", "node", ip); err != nil {
		return err
	}
	if _, _, err := ExecRetry(env.TalosctlBin, "--talosconfig", env.TalosConfig, "bootstrap"); err != nil {
		return err
	}
	if _, _, err := ExecRetry(env.TalosctlBin, "--talosconfig", env.TalosConfig, "kubeconfig", env.ClusterDir); err != nil {
		return err
	}
	if _, _, err := ExecRetry(env.KubectlBin, "--kubeconfig", env.KubeConfig, "get", "nodes"); err != nil {
		return err
	}
	return nil
}

func (env *TestEnv) StopTalosCluster() error {
	fmt.Printf("Stopping talos cluster\n")
	if _, _, err := ExecRetry(env.TalosctlBin, "--talosconfig", env.TalosConfig, "reset", "--reboot", "--graceful=false", "--system-labels-to-wipe", "STATE", "--system-labels-to-wipe", "EPHEMERAL"); err != nil {
		return err
	}
	return nil
}

package test

import (
	_ "embed"
	"fmt"
	"os"
	"path"
)

//go:embed cluster-patch.yaml
var talosPatchYaml string

var (
	clusterDir  = path.Join(BaseDir, "cluster")
	TalosConfig = path.Join(clusterDir, "talosconfig")
	KubeConfig  = path.Join(clusterDir, "kubeconfig")
)

func StartTalosCluster(ip string, name string) error {
	fmt.Printf("Starting talos cluster\n")
	if _, err := os.Stat(TalosConfig); err != nil && os.IsNotExist(err) {
		if _, _, err := Exec(TalosctlBin, "gen", "config", name, "https://"+ip+":6443", "--config-patch", talosPatchYaml, "--output-dir", clusterDir); err != nil {
			return err
		}
	}
	if _, _, err := ExecRetry(TalosctlBin, "apply-config", "--insecure", "--nodes", ip, "--file", path.Join(clusterDir, "controlplane.yaml")); err != nil {
		return err
	}
	if _, _, err := ExecRetry(TalosctlBin, "--talosconfig", TalosConfig, "config", "endpoint", ip); err != nil {
		return err
	}
	if _, _, err := ExecRetry(TalosctlBin, "--talosconfig", TalosConfig, "config", "node", ip); err != nil {
		return err
	}
	if _, _, err := ExecRetry(TalosctlBin, "--talosconfig", TalosConfig, "bootstrap"); err != nil {
		return err
	}
	if _, _, err := ExecRetry(TalosctlBin, "--talosconfig", TalosConfig, "kubeconfig", clusterDir); err != nil {
		return err
	}
	if _, _, err := ExecRetry(KubectlBin, "--kubeconfig", KubeConfig, "get", "nodes"); err != nil {
		return err
	}
	return nil
}

func StopTalosCluster() error {
	fmt.Printf("Stopping talos cluster\n")
	if _, _, err := ExecRetry(TalosctlBin, "--talosconfig", TalosConfig, "reset", "--reboot", "--graceful=false", "--system-labels-to-wipe", "STATE", "--system-labels-to-wipe", "EPHEMERAL"); err != nil {
		return err
	}
	return nil
}

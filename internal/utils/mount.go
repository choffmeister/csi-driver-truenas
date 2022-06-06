package utils

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
	"k8s.io/utils/exec"
)

var _ logr.LogSink = (*logSink)(nil)

type logSink struct{}

func (s logSink) Init(info logr.RuntimeInfo) {}

func (s logSink) Enabled(level int) bool {
	return true
}

func (s logSink) Info(level int, msg string, keysAndValues ...interface{}) {
	Info.Printf(msg, keysAndValues...)
}

func (s logSink) Error(err error, msg string, keysAndValues ...interface{}) {
	Error.Printf(msg, keysAndValues...)
}

func (s logSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return s
}

func (s logSink) WithName(name string) logr.LogSink {
	return s
}

func init() {
	klog.SetLogger(logr.New(logSink{}))
}

type MountUtils struct {
	exec               *exec.Interface
	mount              *mount.Interface
	safeFormatAndMount *mount.SafeFormatAndMount
	resizeFs           *mount.ResizeFs
}

func NewMountUtils() *MountUtils {
	execInstance := exec.New()
	mountInstance := mount.New("")
	return &MountUtils{
		exec:  &execInstance,
		mount: &mountInstance,
		safeFormatAndMount: &mount.SafeFormatAndMount{
			Interface: mountInstance,
			Exec:      execInstance,
		},
		resizeFs: mount.NewResizeFs(execInstance),
	}
}

func (u *MountUtils) FormatAndMountDevice(device string, target string, fstype string) error {
	Info.Printf("Mounting device %s to %s\n", device, target)
	if err := os.MkdirAll(target, 0o775); err != nil {
		return fmt.Errorf("unable to create mount target path: %w", err)
	}
	return u.safeFormatAndMount.FormatAndMount(device, target, fstype, []string{})
}

func (u *MountUtils) UnmountDevice(target string) error {
	Info.Printf("Unmounting device at %s\n", target)
	return u.safeFormatAndMount.Unmount(target)
}

func (u *MountUtils) ResizeDevice(device string, target string) error {
	_, err := u.resizeFs.Resize(device, target)
	return err
}

func (u *MountUtils) GetDeviceNameFromMount(path string) (string, int, error) {
	mounter := mount.New("")
	return mount.GetDeviceNameFromMount(mounter, path)
}

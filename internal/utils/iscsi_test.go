package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ISCSIUtils_GenerateDeviceName(t *testing.T) {
	iscsiUtils := NewISCSIUtils()
	deviceName, err := iscsiUtils.GenerateDeviceName("1.2.3.4", 3260, "iqn.2005-10.org.freenas.ctl:pvc-00000001-0002-0003-0004-00000000005")
	assert.NoError(t, err)
	assert.Equal(t, "/dev/disk/by-path/ip-1.2.3.4:3260-iscsi-iqn.2005-10.org.freenas.ctl:pvc-00000001-0002-0003-0004-00000000005-lun-0", deviceName)
}

func Test_ISCSIUtils_ParseDeviceName(t *testing.T) {
	iscsiUtils := NewISCSIUtils()
	deviceName := "/dev/disk/by-path/ip-1.2.3.4:3260-iscsi-iqn.2005-10.org.freenas.ctl:pvc-00000001-0002-0003-0004-00000000005-lun-0"
	portalIP, portalPort, target, err := iscsiUtils.ParseDeviceName(deviceName)
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3.4", portalIP)
	assert.Equal(t, 3260, portalPort)
	assert.Equal(t, "iqn.2005-10.org.freenas.ctl:pvc-00000001-0002-0003-0004-00000000005", target)

	_, _, _, err = iscsiUtils.ParseDeviceName("/unknown")
	assert.Error(t, err)
}

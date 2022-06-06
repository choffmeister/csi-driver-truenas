package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ISCSIUtils struct {
	mutex sync.Mutex
}

func NewISCSIUtils() *ISCSIUtils {
	return &ISCSIUtils{}
}

func (u *ISCSIUtils) GenerateDeviceName(portalIP string, portalPort int, target string) (string, error) {
	// in this environment the LUN will always be 0, as target names are derived from uniquely generated
	// PV names and so TrueNAS always assigns the LUN 0
	lun := 0
	return fmt.Sprintf("/dev/disk/by-path/ip-%s:%d-iscsi-%s-lun-%d", portalIP, portalPort, target, lun), nil
}

func (u *ISCSIUtils) ParseDeviceName(deviceName string) (string, int, string, error) {
	re := regexp.MustCompile(`^/dev/disk/by-path/ip-(?P<portalIP>[^:]+):(?P<portalPort>\d+)-iscsi-(?P<target>.+)-lun-(?P<lun>\d+)$`)
	m := re.FindStringSubmatch(deviceName)

	if len(m) == 0 {
		return "", 0, "", fmt.Errorf("device name %s is not a valid iscsi by-path device", deviceName)
	}

	portalIP := m[1]
	portalPort, err := strconv.Atoi(m[2])
	if err != nil {
		return "", 0, "", err
	}
	portalTarget := m[3]

	return portalIP, portalPort, portalTarget, nil
}

func (u *ISCSIUtils) Login(portalIP string, portalPort int, target string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	portalAddress := fmt.Sprintf("%s:%d", portalIP, portalPort)
	Info.Printf("Starting iscsi session for %s@%s\n", target, portalAddress)

	if _, _, err := Command("iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", portalAddress); err != nil {
		Command("iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", portalAddress, "-o", "delete")
		return fmt.Errorf("executing iscsiadm failed: %w", err)
	}
	if _, code, err := Command("iscsiadm", "-m", "node", "-T", target, "-p", portalAddress, "--login"); code == 15 {
		Warn.Printf("There already exists an iscsi session for %s@%s already exists\n", target, portalAddress)
	} else if err != nil {
		Command("iscsiadm", "-m", "node", "-T", target, "-p", portalAddress, "-o", "delete")
		return fmt.Errorf("executing iscsiadm failed: %w", err)
	}

	hostNumber, err := u.GetHostNumber(target, portalAddress)
	if err != nil {
		return fmt.Errorf("unable to get scsi host number for target %s on portal %s: %w", target, portalAddress, err)
	}

	// Scan the iSCSI bus for the LUN
	if err := u.ScanLUN(hostNumber, 0); err != nil {
		return fmt.Errorf("unable to scan lun %d on scsi host %d", 0, hostNumber)
	}

	return nil
}

func (u *ISCSIUtils) Logout(portalIP string, portalPort int, target string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	portalAddress := fmt.Sprintf("%s:%d", portalIP, portalPort)
	Info.Printf("Stopping session for %s@%s\n", target, portalAddress)

	if _, code, err := Command("iscsiadm", "-m", "node", "-T", target, "-p", portalAddress, "--logout"); code == 21 {
		Warn.Printf("No iscsi session for %s@%s exists\n", target, portalAddress)
	} else if err != nil {
		return fmt.Errorf("executing iscsiadm failed: %w", err)
	}

	return nil
}

func (u *ISCSIUtils) Rescan(target string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	Info.Printf("Rescanning session for %s\n", target)

	if _, _, err := Command("iscsiadm", "-m", "node", "--targetname", target, "-R"); err != nil {
		return fmt.Errorf("executing iscsiadm failed: %w", err)
	}

	return nil
}

func (u *ISCSIUtils) GetHostNumber(target string, portalAddress string) (int, error) {
	maxAttempts := 3
	attempt := 0
	for {
		portalHostMap, err := u.GetISCSIPortalHostMapForTarget(target)
		if err != nil {
			if attempt < maxAttempts {
				time.Sleep(time.Second)
				attempt++
				continue
			} else {
				return 0, fmt.Errorf("getting iscsi portal host map failed failed: %w", err)
			}
		}
		hostNumber, loggedIn := portalHostMap[portalAddress]
		if !loggedIn {
			if attempt < maxAttempts {
				attempt++
				time.Sleep(time.Second)
				continue
			} else {
				return 0, fmt.Errorf("could not get scsi host number for portal %s after logging in: %w", portalAddress, err)
			}
		}

		return hostNumber, nil
	}
}

// https://github.com/kubernetes/kubernetes
func (u *ISCSIUtils) ScanLUN(hostNumber int, lunNumber int) error {
	filename := fmt.Sprintf("/host/sys/class/scsi_host/host%d/scan", hostNumber)
	file, err := os.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	// Channel/Target are always 0 for iSCSI
	scanCmd := fmt.Sprintf("0 0 %d", lunNumber)
	if written, err := file.WriteString(scanCmd); err != nil {
		return err
	} else if written == 0 {
		return fmt.Errorf("no data written to file: %s", filename)
	}

	Info.Printf("Scanned SCSI host %d LUN %d\n", hostNumber, lunNumber)
	return nil
}

// https://github.com/kubernetes/kubernetes
// GetISCSIPortalHostMapForTarget given a target iqn, find all the scsi hosts logged into
// that target. Returns a map of iSCSI portals (string) to SCSI host numbers (integers).
// For example: {
//    "192.168.30.7:3260": 2,
//    "192.168.30.8:3260": 3,
// }
func (u *ISCSIUtils) GetISCSIPortalHostMapForTarget(targetIqn string) (map[string]int, error) {
	portalHostMap := make(map[string]int)

	// Iterate over all the iSCSI hosts in sysfs
	sysPath := "/sys/class/iscsi_host"
	hostDirs, err := ioutil.ReadDir(sysPath)
	if err != nil {
		if os.IsNotExist(err) {
			return portalHostMap, nil
		}
		return nil, err
	}
	for _, hostDir := range hostDirs {
		// iSCSI hosts are always of the format "host%d"
		// See drivers/scsi/hosts.c in Linux
		hostName := hostDir.Name()
		if !strings.HasPrefix(hostName, "host") {
			continue
		}
		hostNumber, err := strconv.Atoi(strings.TrimPrefix(hostName, "host"))
		if err != nil {
			Error.Printf("Could not get number from iSCSI host: %s\n", hostName)
			continue
		}

		// Iterate over the children of the iscsi_host device
		// We are looking for the associated session
		devicePath := sysPath + "/" + hostName + "/device"
		deviceDirs, err := ioutil.ReadDir(devicePath)
		if err != nil {
			return nil, err
		}
		for _, deviceDir := range deviceDirs {
			// Skip over files that aren't the session
			// Sessions are of the format "session%u"
			// See drivers/scsi/scsi_transport_iscsi.c in Linux
			sessionName := deviceDir.Name()
			if !strings.HasPrefix(sessionName, "session") {
				continue
			}

			sessionPath := devicePath + "/" + sessionName

			// Read the target name for the iSCSI session
			targetNamePath := sessionPath + "/iscsi_session/" + sessionName + "/targetname"
			targetName, err := ioutil.ReadFile(targetNamePath)
			if err != nil {
				Info.Printf("Failed to process session %s, assuming this session is unavailable: %s\n", sessionName, err)
				continue
			}

			// Ignore hosts that don't matchthe target we were looking for.
			if strings.TrimSpace(string(targetName)) != targetIqn {
				continue
			}

			// Iterate over the children of the iSCSI session looking
			// for the iSCSI connection.
			dirs2, err := ioutil.ReadDir(sessionPath)
			if err != nil {
				Info.Printf("Failed to process session %s, assuming this session is unavailable: %s\n", sessionName, err)
				continue
			}
			for _, dir2 := range dirs2 {
				// Skip over files that aren't the connection
				// Connections are of the format "connection%d:%u"
				// See drivers/scsi/scsi_transport_iscsi.c in Linux
				dirName := dir2.Name()
				if !strings.HasPrefix(dirName, "connection") {
					continue
				}

				connectionPath := sessionPath + "/" + dirName + "/iscsi_connection/" + dirName

				// Read the current and persistent portal information for the connection.
				addrPath := connectionPath + "/address"
				addr, err := ioutil.ReadFile(addrPath)
				if err != nil {
					Info.Printf("Failed to process connection %s, assuming this connection is unavailable: %s\n", dirName, err)
					continue
				}

				portPath := connectionPath + "/port"
				port, err := ioutil.ReadFile(portPath)
				if err != nil {
					Info.Printf("Failed to process connection %s, assuming this connection is unavailable: %s\n", dirName, err)
					continue
				}

				persistentAddrPath := connectionPath + "/persistent_address"
				persistentAddr, err := ioutil.ReadFile(persistentAddrPath)
				if err != nil {
					Info.Printf("Failed to process connection %s, assuming this connection is unavailable: %s\n", dirName, err)
					continue
				}

				persistentPortPath := connectionPath + "/persistent_port"
				persistentPort, err := ioutil.ReadFile(persistentPortPath)
				if err != nil {
					Info.Printf("Failed to process connection %s, assuming this connection is unavailable: %s\n", dirName, err)
					continue
				}

				// Add entries to the map for both the current and persistent portals
				// pointing to the SCSI host for those connections
				portal := strings.TrimSpace(string(addr)) + ":" +
					strings.TrimSpace(string(port))
				portalHostMap[portal] = hostNumber

				persistentPortal := strings.TrimSpace(string(persistentAddr)) + ":" +
					strings.TrimSpace(string(persistentPort))
				portalHostMap[persistentPortal] = hostNumber
			}
		}
	}

	return portalHostMap, nil
}

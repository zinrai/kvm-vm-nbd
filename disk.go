package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func connectNBD(diskPath, nbdDevice string) error {
	cmd := exec.Command("qemu-nbd", "--connect="+nbdDevice, diskPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect NBD: %v, output: %s", err, output)
	}
	fmt.Printf("NBD device %s connected\n", nbdDevice)

	// Wait for partition device nodes to appear under /dev.
	time.Sleep(3 * time.Second)
	return nil
}

func disconnectNBD(nbdDevice string) error {
	cmd := exec.Command("qemu-nbd", "--disconnect", nbdDevice)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to disconnect NBD: %v, output: %s", err, output)
	}
	fmt.Printf("NBD device %s disconnected\n", nbdDevice)
	return nil
}

func listNBDPartitions(nbdDevice string) ([]string, error) {
	matches, err := filepath.Glob(nbdDevice + "p*")
	if err != nil {
		return nil, err
	}
	sort.Slice(matches, func(i, j int) bool {
		ni, _ := strconv.Atoi(partitionNumber(matches[i], nbdDevice))
		nj, _ := strconv.Atoi(partitionNumber(matches[j], nbdDevice))
		return ni < nj
	})
	return matches, nil
}

func partitionNumber(partition, nbdDevice string) string {
	return strings.TrimPrefix(partition, nbdDevice+"p")
}

func getFSType(partition string) (string, error) {
	out, err := exec.Command("blkid", "-o", "value", "-s", "TYPE", partition).Output()
	if err != nil {
		// blkid exits non-zero when the FS is unrecognized; treat as empty.
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func getUUID(partition string) (string, error) {
	out, err := exec.Command("blkid", "-o", "value", "-s", "UUID", partition).Output()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

// findNBDByDiskPath returns the /dev/nbdN whose qemu-nbd command line
// references diskPath, or "" if none is found. The mapping is discovered
// by reading /sys/block/nbd*/pid and inspecting /proc/<pid>/cmdline.
func findNBDByDiskPath(diskPath string) (string, error) {
	pidPaths, err := filepath.Glob("/sys/block/nbd*/pid")
	if err != nil {
		return "", err
	}
	for _, p := range pidPaths {
		pidData, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		pid := strings.TrimSpace(string(pidData))
		if pid == "" {
			continue
		}
		cmdlineData, err := os.ReadFile("/proc/" + pid + "/cmdline")
		if err != nil {
			continue
		}
		args := strings.Split(strings.TrimRight(string(cmdlineData), "\x00"), "\x00")
		for _, a := range args {
			if a == diskPath {
				name := filepath.Base(filepath.Dir(p))
				return "/dev/" + name, nil
			}
		}
	}
	return "", nil
}

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func getShutOffVMs() ([]string, error) {
	cmd := exec.Command("virsh", "list", "--state-shutoff", "--name")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting shut off VMs: %v", err)
	}

	vms := strings.Split(strings.TrimSpace(string(output)), "\n")
	var nonEmptyVMs []string
	for _, vm := range vms {
		if vm != "" {
			nonEmptyVMs = append(nonEmptyVMs, vm)
		}
	}

	return nonEmptyVMs, nil
}

func getDiskPaths(vmName string) ([]string, error) {
	cmd := exec.Command("virsh", "dumpxml", vmName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting VM XML: %v", err)
	}

	var diskPaths []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "source file=") {
			path := strings.Split(line, "'")[1]
			diskPaths = append(diskPaths, path)
		}
	}

	return diskPaths, nil
}

func resolveDiskPath(vmName string, diskIdx int) (string, error) {
	diskPaths, err := getDiskPaths(vmName)
	if err != nil {
		return "", err
	}
	if len(diskPaths) == 0 {
		return "", fmt.Errorf("no disks found for VM %q", vmName)
	}
	if diskIdx < 1 || diskIdx > len(diskPaths) {
		return "", fmt.Errorf("--disk %d out of range (VM %q has %d disk(s))", diskIdx, vmName, len(diskPaths))
	}
	return diskPaths[diskIdx-1], nil
}

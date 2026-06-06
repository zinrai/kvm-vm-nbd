package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func cmdUmount(args []string) {
	fs := flag.NewFlagSet("umount", flag.ExitOnError)
	diskIdx := fs.Int("disk", 1, "Disk index (1-based) when the VM has multiple disks")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: kvm-vm-nbd umount [--disk N] <vm-name>

Unmount partitions of the VM's disk image and disconnect the NBD device.
The NBD device is discovered by reading /sys/block/nbd*/pid and inspecting
the qemu-nbd command line via /proc/<pid>/cmdline.

`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(2)
	}
	vmName := fs.Arg(0)

	diskPath, err := resolveDiskPath(vmName, *diskIdx)
	if err != nil {
		log.Fatalf("%v", err)
	}

	nbdDevice, err := findNBDByDiskPath(diskPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if nbdDevice == "" {
		log.Fatalf("no NBD device is attached to %s", diskPath)
	}

	mounts, err := getMountedPartitions(nbdDevice)
	if err != nil {
		log.Fatalf("%v", err)
	}
	for _, mp := range mounts {
		if err := unmountPartition(mp); err != nil {
			fmt.Fprintf(os.Stderr, "error unmounting %s: %v\n", mp, err)
		} else {
			fmt.Printf("Unmounted %s\n", mp)
		}
	}

	if err := disconnectNBD(nbdDevice); err != nil {
		log.Fatalf("%v", err)
	}
}

func unmountPartition(mountPoint string) error {
	cmd := exec.Command("umount", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount %s: %v", mountPoint, err)
	}
	return nil
}

func getMountedPartitions(nbdDevice string) ([]string, error) {
	out, err := exec.Command("mount").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read mount table: %v", err)
	}

	var mps []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, nbdDevice) {
			fields := strings.Fields(line)
			if len(fields) > 2 {
				mps = append(mps, fields[2])
			}
		}
	}
	return mps, nil
}

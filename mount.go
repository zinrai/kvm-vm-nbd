package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
)

func cmdMount(args []string) {
	fs := flag.NewFlagSet("mount", flag.ExitOnError)
	diskIdx := fs.Int("disk", 1, "Disk index (1-based) when the VM has multiple disks")
	partNum := fs.Int("partition", 1, "Partition number (1-based)")
	nbd := fs.String("nbd", "/dev/nbd0", "NBD device to attach the disk image to")
	mountPoint := fs.String("mount", "/mnt/vm_partition", "Mount point directory")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: kvm-vm-nbd mount [--disk N] [--partition N] [--nbd /dev/nbdN] [--mount PATH] <vm-name>

Attach the VM's disk image to an NBD device and mount the specified partition.

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

	vms, err := getShutOffVMs()
	if err != nil {
		log.Fatalf("%v", err)
	}
	if !slices.Contains(vms, vmName) {
		log.Fatalf("VM %q is not in shut-off state", vmName)
	}

	diskPath, err := resolveDiskPath(vmName, *diskIdx)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if err := connectNBD(diskPath, *nbd); err != nil {
		log.Fatalf("%v", err)
	}

	partition := fmt.Sprintf("%sp%d", *nbd, *partNum)
	if _, err := os.Stat(partition); err != nil {
		disconnectNBD(*nbd)
		log.Fatalf("partition %s not found: %v", partition, err)
	}

	if err := mountPartition(partition, *mountPoint); err != nil {
		disconnectNBD(*nbd)
		log.Fatalf("%v", err)
	}
}

func mountPartition(partition, mountPoint string) error {
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	fsType, err := getFSType(partition)
	if err != nil {
		return fmt.Errorf("failed to detect file system type of %s: %v", partition, err)
	}
	if fsType == "" {
		return fmt.Errorf("unrecognized file system on %s", partition)
	}

	cmd := exec.Command("mount", "-t", fsType, partition, mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mount partition: %v, output: %s", err, string(output))
	}

	fmt.Printf("Successfully mounted %s to %s\n", partition, mountPoint)
	return nil
}

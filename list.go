package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
)

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	nbd := fs.String("nbd", "/dev/nbd0", "NBD device used for inspection")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: kvm-vm-nbd list [--nbd /dev/nbdN] <vm-name>

List disks and partitions of a shut-off VM by temporarily attaching each
disk image to an NBD device.

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
		log.Fatalf("VM %q is not in shut-off state (only shut-off VMs can be inspected)", vmName)
	}

	diskPaths, err := getDiskPaths(vmName)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if len(diskPaths) == 0 {
		log.Fatalf("no disks found for VM %q", vmName)
	}

	for i, disk := range diskPaths {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("Disk: %s\n", disk)
		if err := inspectDisk(disk, *nbd); err != nil {
			fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		}
	}
}

func inspectDisk(diskPath, nbdDevice string) error {
	if err := connectNBD(diskPath, nbdDevice); err != nil {
		return err
	}
	defer disconnectNBD(nbdDevice)

	partitions, err := listNBDPartitions(nbdDevice)
	if err != nil {
		return err
	}
	if len(partitions) == 0 {
		fmt.Println("  (no partitions)")
		return nil
	}

	fmt.Printf("  %-4s %-14s %-8s %s\n", "NUM", "DEVICE", "FSTYPE", "UUID")
	for _, p := range partitions {
		num := partitionNumber(p, nbdDevice)
		fstype, _ := getFSType(p)
		uuid, _ := getUUID(p)
		if fstype == "" {
			fstype = "-"
		}
		if uuid == "" {
			uuid = "-"
		}
		fmt.Printf("  %-4s %-14s %-8s %s\n", num, p, fstype, uuid)
	}
	return nil
}

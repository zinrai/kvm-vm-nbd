package main

import (
	"fmt"
	"log"
	"os"
)

const usageText = `Usage: kvm-vm-nbd <command> [arguments]

Commands:
  list     List disks and partitions of a shut-off VM
  mount    Mount a partition of the VM's disk image
  umount   Unmount partitions and disconnect an NBD device

Run 'kvm-vm-nbd <command> -h' for command-specific flags.
`

func main() {
	log.SetFlags(0)
	log.SetPrefix("kvm-vm-nbd: ")

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "list":
		cmdList(os.Args[2:])
	case "mount":
		cmdMount(os.Args[2:])
	case "umount":
		cmdUmount(os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usageText)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", os.Args[1], usageText)
		os.Exit(2)
	}
}

# kvm-vm-nbd

`kvm-vm-nbd` is a command-line tool for mounting partitions of KVM virtual machine disk images via NBD (Network Block Device). All subcommands take a libvirt VM name as their positional argument.

**LVM is not supported.**

## Prerequisites

- KVM and QEMU installed on the host
- `qemu-nbd`, `blkid`, `mount`, `umount`, `virsh`
- Root or sudo access (required for NBD attach and mount operations)

## Subcommands

```
list <vm-name>      List disks and partitions of a shut-off VM
mount <vm-name>     Mount a partition of the VM's disk image
umount <vm-name>    Unmount partitions and disconnect the NBD device
```

Run `kvm-vm-nbd <command> -h` for command-specific flags.

`mount` selects which partition to mount via `--partition N` (1-based, default `1`). When a VM has multiple disks, `--disk N` (1-based, default `1`) selects which disk `mount` and `umount` operate on.

## Usage

Inspect a shut-off VM's disks and partitions:

```
$ sudo kvm-vm-nbd list bookworm64
Disk: /var/lib/libvirt/images/bookworm64.qcow2
NBD device /dev/nbd0 connected
  NUM  DEVICE         FSTYPE   UUID
  1    /dev/nbd0p1    ext4     c6d5ebcc-6882-47d8-a88a-fb2b01de4ed6
  5    /dev/nbd0p5    swap     bb7d71f0-e8fb-4f46-8d46-1430a17e647a
NBD device /dev/nbd0 disconnected
```

Mount the VM's disk. With no flags, this mounts `--disk 1`'s `--partition 1` at the default mount point:

```
$ sudo kvm-vm-nbd mount bookworm64
NBD device /dev/nbd0 connected
Successfully mounted /dev/nbd0p1 to /mnt/vm_partition
```

Unmount and disconnect — the NBD device is discovered automatically by reading `/sys/block/nbd*/pid` and inspecting the qemu-nbd command line via `/proc/<pid>/cmdline`:

```
$ sudo kvm-vm-nbd umount bookworm64
Unmounted /mnt/vm_partition
NBD device /dev/nbd0 disconnected
```

Mount a specific partition on a different NBD device and mount point (e.g. to mount multiple VMs in parallel):

```
$ sudo kvm-vm-nbd mount bookworm64 --partition 1 --nbd /dev/nbd2 --mount /mnt/work
$ sudo kvm-vm-nbd umount bookworm64
```

The number passed to `--partition` is the partition table index — the `1` in `/dev/nbd0p1` — and matches the `NUM` column shown by `list`.

## Notes

- `list` and `mount` require the VM to be shut off.
- `list` temporarily attaches each disk to the NBD device passed via `--nbd`; that device must not already be in use.
- `umount` does not require the VM to be shut off, so it can clean up after the VM has been started again.
- `umount` performs unmount and `qemu-nbd --disconnect` as a single step.
- Modifying a mounted disk image can corrupt the guest filesystem — proceed with care.

## License

This project is licensed under the [MIT License](./LICENSE).

# fmount
Mount various disk images for forensic read-only processing.

```console
go install go.foxforensics.dev/fmount@latest
```

## Usage
```console
# fmount [-ruszqhv] [-H CRC32|MD5|SHA1|SHA256] [-V SUM] [-B KEY] [-D DIR] IMAGE
```

Available options:

- `-D` Mount point
- `-B` BitLocker key
- `-H` Hash algorithm
- `-V` Verify hash sum
- `-r` Recovery key ids
- `-u` Unmount image
- `-s` System partition only
- `-z` Unzip image
- `-q` Quiet mode
- `-h` Show usage
- `-v` Show version

Required system commands:

- [dislocker](https://github.com/Aorimn/dislocker)
- [qemu-nbd](https://www.qemu.org/docs/master/tools/qemu-nbd.html)
- [lsblk](https://man7.org/linux/man-pages/man8/lsblk.8.html)
- [mount](https://man7.org/linux/man-pages/man8/mount.8.html)
- [umount](https://man7.org/linux/man-pages/man8/umount.8.html)

## Formats
Supported image types on Linux host systems:

- [vdi](https://forensics.wiki/virtual_disk_image_%28vdi%29/)
- [vpc](https://cloud.ibm.com/docs/vpc?topic=vpc-planning-custom-images)
- [vhdx](https://forensics.wiki/virtual_hard_disk_%28vhd%29/)
- [vmdk](https://forensics.wiki/vmware_virtual_disk_format_%28vmdk%29/)
- [parallels](https://github.com/libyal/libphdi/blob/main/documentation/Parallels%20Hard%20Disk%20image%20format.asciidoc)
- [qcow2](https://forensics.wiki/qcow_image_format/)
- [qcow](https://forensics.wiki/qcow_image_format/)
- [raw](https://forensics.wiki/raw_image_format/)

## License
Released under the [MIT License](LICENSE.md).
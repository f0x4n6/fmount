package fmount

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/f0x4n6/futils/pkg/hash"
	"github.com/f0x4n6/futils/pkg/sys"
	"github.com/f0x4n6/futils/pkg/zip"

	"github.com/f0x4n6/fmount/internal/posix"
)

const (
	Version = "0.5.5"
)

func Mount(img, mnt, key string, so bool, xargs []string) (parts []string, err error) {

	// create symlink directory
	if err = os.MkdirAll(posix.SymlinkPath, sys.MODE_DIR); err != nil {
		return
	}

	// create mount point
	if mnt, err = posix.CreateImageMount(img, mnt); err != nil {
		return
	}

	// ensure kernel module is loaded
	if err = posix.EnsureMod(posix.QemuParts); err != nil {
		return
	}

	// attach image as network block device
	if err = posix.QemuAttach(posix.QemuDev, img, xargs); err != nil {
		return
	}

	// create image symlink directory to track image relations
	if err = posix.CreateImageSymlink(img, posix.QemuDev); err != nil {
		return
	}

	// get partition network block devices
	nbdps, err := posix.PartDevs(posix.QemuDev)

	if err != nil {
		return
	}

	// handle found partitions
	for i, nbdp := range nbdps {
		dev := posix.Dev(nbdp)

		// check if partition is bootable
		sp, err := posix.IsBootable(dev)

		if err != nil {
			sys.Error(err)
			continue
		}

		// if all or bootable
		if !so || sp {

			// create partition mount point
			mntp, err := posix.CreateDirf(mnt, "p%d", i+1)

			if err != nil {
				sys.Error(err)
				continue
			}

			// check if partition is encrypted
			is, err := posix.IsEncrypted(dev)

			if err != nil {
				sys.Error(err)
				continue
			}

			// if encrypted
			if is {

				// check if key given
				if len(key) == 0 {
					sys.Error("no key given")
					continue
				}

				// create fuse mount point
				mntf, err := posix.CreateDirf(mnt, "p%d-fuse", i+1)

				if err != nil {
					sys.Error(err)
					continue
				}

				// mount to be decrypted partition as fuse
				if err = posix.DislockerFuse(dev, key, mntf); err != nil {
					sys.Error(err)
					continue
				}

				// create symlink to track device relations
				if err = posix.CreateSymlink(nbdp, mntf); err != nil {
					sys.Error(err)
					continue
				}

				// overwrite device to be mounted
				dev = filepath.Join(mntf, posix.DislockerDev)
			}

			// mount device
			if posix.Mount(dev, mntp, is) != nil {
				sys.Error(err)
				continue
			}

			// report progress
			mntp, err = filepath.Abs(mntp)

			if err != nil {
				sys.Error(err)
				continue
			}

			if sys.Progress != nil {
				sys.Progress(mntp)
			}

			parts = append(parts, mntp)
		}
	}

	return parts, nil
}

func Unmount(img string) (err error) {
	img, err = filepath.Abs(img)

	if err != nil {
		return
	}

	// get network block devices associated with image
	nbds, err := posix.BlockDevs(img)

	if err != nil {
		return
	}

	// handle found network block devices
	for _, nbd := range nbds {

		// get partition devices
		nbdps, err := posix.PartDevs(nbd)

		if err != nil {
			sys.Error(err)
			continue
		}

		// get mount points of device
		mnts, err := posix.Mounts(nbd)

		if err != nil {
			sys.Error(err)
			continue
		}

		// handle found partitions
		for _, nbdp := range nbdps {
			dev := posix.Dev(nbdp)

			// check if partition is encrypted
			is, err := posix.IsEncrypted(dev)

			if err != nil {
				sys.Error(err)
				continue
			}

			// if encrypted
			if is {

				// follow symlink and get partition mount points
				src, err := posix.FollowSymlink(nbdp)

				if err != nil {
					sys.Error(err)
					continue
				}

				mntf := filepath.Dir(src)
				mntp := posix.FromFuse(src)
				mnts = append(mnts, mntp)

				// unmount partition mount point
				if err = posix.UmountDir(mntp); err != nil {
					sys.Error(err)
					continue
				}

				// unmount fuse mount point
				if err = posix.UmountDir(mntf); err != nil {
					sys.Error(err)
					continue
				}

				// detach partition network block device
				if err = posix.QemuDetach(dev); err != nil {
					sys.Error(err)
					continue
				}

				// remove symlink
				if err = posix.RemoveSymlink(nbdp); err != nil {
					sys.Error(err)
					continue
				}
			} else {

				// unmount partition network block device
				if err = posix.UmountDev(dev); err != nil {
					sys.Error(err)
					continue
				}

				// detach partition network block device
				if err = posix.QemuDetach(dev); err != nil {
					sys.Error(err)
					continue
				}
			}
		}

		// remove image symlink directory
		dir := filepath.Join(posix.SymlinkPath, filepath.Base(img))

		if err = os.RemoveAll(dir); err != nil {
			sys.Error(err)
		}

		// remove empty mount points
		for _, mnt := range mnts {
			err = posix.RemoveDirs(filepath.Dir(mnt))

			if err != nil {
				sys.Error(err)
				continue
			}
		}
	}

	return nil
}

func KeyIds(img string, xargs []string) (ids []string, err error) {
	if err = posix.EnsureMod(posix.QemuParts); err != nil {
		return
	}

	if err = posix.QemuAttach(posix.QemuDev, img, xargs); err != nil {
		return
	}

	nbdps, err := posix.PartDevs(posix.QemuDev)

	if err != nil {
		return
	}

	for _, nbdp := range nbdps {
		dev := posix.Dev(nbdp)

		idps, err := posix.DislockerInfo(dev)

		if err != nil {
			sys.Error(err)
			continue
		}

		if sys.Progress != nil {
			for _, idp := range idps {
				sys.Progress(idp)
			}
		}

		ids = append(ids, idps...)

		if err = posix.QemuDetach(dev); err != nil {
			sys.Error(err)
		}
	}

	return
}

func Extract(img string) (p string, err error) {
	i, err := zip.Index(img)

	if err != nil {
		return
	}

	if len(i) > 1 {
		err = errors.New("more than one file")
		return
	}

	dir := filepath.Dir(img)

	p = filepath.Join(dir, i[0])

	if _, err = os.Stat(p); !os.IsNotExist(err) {
		err = errors.New("file already exists")
		return
	}

	if err = zip.Unzip(img, dir); err != nil {
		return
	}

	return
}

func Verify(img, algo, sum string) (ok bool, err error) {
	b, err := hash.Sum(img, algo)

	if err != nil {
		return
	}

	ok = fmt.Sprintf("%x", b) == strings.ToLower(sum)

	return
}

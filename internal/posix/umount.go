package posix

import (
	"github.com/f0x4n6/futils/pkg/sys"
)

func UmountDir(dir string) (err error) {
	_, err = sys.StdCall("umount", "-R", dir)

	return
}

func UmountDev(dev string) (err error) {
	_, err = sys.StdCall("umount", "-A", dev)

	return
}

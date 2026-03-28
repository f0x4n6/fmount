package posix

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.foxforensics.dev/futils/pkg/sys"
)

const (
	SymlinkPath = "/tmp/fmount"
)

func Dev(loop string) string {
	return filepath.Join("/dev", loop)
}

func FromFuse(dev string) string {
	return dev[:len(dev)-len("-fuse/"+DislockerDev)]
}

func BaseFile(name string) string {
	b := filepath.Base(name)

	return strings.TrimSuffix(b, filepath.Ext(b))
}

func EnsureMod(max int) (err error) {
	ls, err := ModList("nbd")

	if err != nil {
		return
	}

	for _, l := range ls {
		if strings.HasPrefix(l, "nbd ") {
			return
		}
	}

	return ModLoad("nbd", fmt.Sprintf("max_part=%d", max))
}

func BlockDevs(img string) (nbds []string, err error) {
	dir := filepath.Join(SymlinkPath, filepath.Base(img))

	ff, err := os.ReadDir(dir)

	if err != nil {
		return
	}

	for _, f := range ff {
		if f.Type()&os.ModeSymlink != os.ModeSymlink {
			continue
		}

		lnk := filepath.Join(dir, f.Name())

		src, err := filepath.EvalSymlinks(lnk)

		if err != nil {
			sys.Error(err)
			continue
		}

		nbds = append(nbds, src)
	}

	if len(nbds) == 0 {
		err = errors.New("no devices found")
		return
	}

	return
}

func PartDevs(dev string) (ps []string, err error) {
	ps, err = LsBlk(dev, "name")

	if err != nil {
		return
	}

	if len(ps) <= 1 {
		err = errors.New("no partitions found")
		return
	}

	ps = ps[1:] // skip root device

	return
}

func Mounts(dev string) (mnts []string, err error) {
	return LsBlk(dev, "mountpoints")
}

func IsBootable(dev string) (is bool, err error) {
	f, err := os.Open(dev)

	if err != nil {
		return
	}

	defer func() { _ = f.Close() }()

	b := make([]byte, 512)

	n, err := f.Read(b)

	if err != nil {
		return
	}

	if n != len(b) {
		err = errors.New("could not read sector")
		return
	}

	is = bytes.Equal(b[0x1FE:0x200], []byte{0x55, 0xAA})

	return
}

func IsEncrypted(dev string) (is bool, err error) {
	ft, err := LsBlk(dev, "fstype")

	if err != nil {
		return
	}

	if len(ft) != 1 {
		err = errors.New("type count differs")
		return
	}

	is = ft[0] == "BitLocker"

	return
}

func CreateImageMount(img, mnt string) (dir string, err error) {
	if len(mnt) == 0 {
		mnt = BaseFile(img)
	}

	return mnt, os.MkdirAll(mnt, sys.MODE_DIR)
}

func CreateImageSymlink(img, dev string) (err error) {
	dir := filepath.Join(SymlinkPath, filepath.Base(img))

	if _, err = os.Stat(dir); !os.IsNotExist(err) {
		if err = os.RemoveAll(dir); err != nil {
			sys.Error(err)
		}
	}

	if err = os.MkdirAll(dir, sys.MODE_DIR); err != nil {
		return
	}

	lnk := filepath.Join(dir, filepath.Base(dev))

	return os.Symlink(dev, lnk)
}

func CreateSymlink(dev, mnt string) error {
	src := filepath.Join(mnt, DislockerDev)
	lnk := filepath.Join(SymlinkPath, dev)

	return os.Symlink(src, lnk)
}

func FollowSymlink(dev string) (src string, err error) {
	lnk := filepath.Join(SymlinkPath, dev)

	if _, err = os.Stat(lnk); os.IsNotExist(err) {
		err = errors.New("symlink not found")
		return
	}

	src, err = filepath.EvalSymlinks(lnk)

	return
}

func RemoveSymlink(dev string) error {
	lnk := filepath.Join(SymlinkPath, dev)

	return os.Remove(lnk)
}

func CreateDirf(root, m string, a ...any) (dir string, err error) {
	dir = filepath.Join(root, fmt.Sprintf(m, a...))
	err = os.MkdirAll(dir, sys.MODE_DIR)

	return
}

func RemoveDirs(dir string) (err error) {
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		return
	}

	ff, err := os.ReadDir(dir)

	if err != nil {
		return
	}

	for _, f := range ff {
		if !f.IsDir() {
			continue
		}

		_ = os.Remove(filepath.Join(dir, f.Name()))
	}

	_ = os.Remove(dir)

	return nil
}

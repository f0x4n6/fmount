package posix

import (
	"strings"

	"github.com/f0x4n6/futils/pkg/sys"
)

const (
	DislockerDev = "dislocker-file"
)

func DislockerInfo(dev string) (ids []string, err error) {
	md, err := sys.StdCall("dislocker-metadata", "-V", dev)

	if err != nil {
		return
	}

	ss := strings.Split(md, "\n")

	for _, s := range ss {
		i := strings.Index(s, "Recovery Key GUID")

		if i >= 0 {
			ids = append(ids, s[i+20:i+56])
		}
	}

	return
}

func DislockerFuse(dev, key, dir string) (err error) {
	_, err = sys.StdCall("dislocker-fuse", "-r", "-V", dev, "-p"+key, "--", dir)

	return
}

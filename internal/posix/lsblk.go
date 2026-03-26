package posix

import (
	"strings"

	"github.com/f0x4n6/futils/pkg/sys"
)

func LsBlk(dev, col string) (ls []string, err error) {
	out, err := sys.StdCall("lsblk", "-l", "-n", "-o", col, strings.TrimSpace(dev))

	if err != nil {
		return
	}

	ls = strings.Split(strings.TrimSpace(out), "\n")

	return
}

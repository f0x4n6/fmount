package posix

import (
	"github.com/f0x4n6/futils/pkg/sys"
)

func Mount(dev, dir string, lo bool) (err error) {
	args := []string{"-o", "ro", dev, dir}

	if lo {
		args = append([]string{"-o", "loop"}, args...)
	}

	_, err = sys.StdCall("mount", args...)

	return
}

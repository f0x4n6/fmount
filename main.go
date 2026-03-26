// Mount various disk images for forensic read-only processing.
//
// Usage:
//
//	fmount [-ruszqhv] [-H CRC32|MD5|SHA1|SHA256] [-V SUM] [-B KEY] [-D DIR] IMAGE
//
// The flags are:
//
//	 -D directory
//		The mount point directory.
//	 -B key
//	 	The BitLocker key.
//	 -H algorithm
//	 	The hash algorithm to use.
//	 -V sum
//	 	The hash sum to verify.
//	 -r
//		Search recovery key ids.
//	 -u
//		Unmount image.
//	 -s
//		System partition only.
//	 -z
//		Unzip image.
//	 -q
//		Quiet mode.
//	 -h
//		Show usage.
//	 -v
//		Show version.
//
// The arguments are:
//
//	 image
//		The disk images filename.
package main

import (
	"flag"
	"io"

	"github.com/f0x4n6/futils/pkg/sys"

	"github.com/f0x4n6/fmount/internal/fmount"
)

func main() {
	D := flag.String("D", "", "Mount point")
	B := flag.String("B", "", "BitLocker key")
	H := flag.String("H", "", "Hash algorithm")
	V := flag.String("V", "", "Hash sum")
	r := flag.Bool("r", false, "Recovery key ids")
	s := flag.Bool("s", false, "System partition only")
	u := flag.Bool("u", false, "Unmount image")
	z := flag.Bool("z", false, "Unzip image")
	q := flag.Bool("q", false, "Quiet mode")
	h := flag.Bool("h", false, "Show usage")
	v := flag.Bool("v", false, "Show version")

	flag.CommandLine.SetOutput(io.Discard)
	flag.Parse()

	args, xargs := sys.Args()

	if *v {
		sys.Final("fmount", fmount.Version)
	}

	if *h || len(args) == 0 {
		sys.Usage("fmount [-ruszqhv] [-H CRC32|MD5|SHA1|SHA256] [-V SUM] [-B KEY] [-D DIR] IMAGE")
	}

	if *q {
		sys.Progress = nil
	}

	img := args[0]

	if *z {
		ex, err := fmount.Extract(img)

		if err != nil {
			sys.Fatal(err)
		} else {
			img = ex
		}
	}

	if (len(*H) == 0) != (len(*V) == 0) {
		sys.Fatal("hash algorithm and sum are required")
	}

	if len(*H) > 0 && len(*V) > 0 {
		ok, err := fmount.Verify(img, *H, *V)

		if err != nil {
			sys.Fatal(err)
		}

		if !ok {
			sys.Fatal("hash sum does not match")
		}
	}

	var err error

	if *u {
		err = fmount.Unmount(img)
	} else if *r {
		_, err = fmount.KeyIds(img, xargs)
	} else {
		_, err = fmount.Mount(img, *D, *B, *s, xargs)
	}

	if err != nil {
		sys.Fatal(err)
	}
}

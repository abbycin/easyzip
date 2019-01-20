package main

import (
	"./easyzip"
	"fmt"
	"os"
)

func main() {
	help := func() {
		fmt.Printf("%s [zip|unzip] src dst\n%s zipfiles dst file1 file2...\n", os.Args[0], os.Args[0])
		os.Exit(1)
	}
	if len(os.Args) < 4 {
		help()
	}
	z := easyzip.NewZip(true)

	cmd := os.Args[1]
	src, dst := os.Args[2], os.Args[3]
	var err error
	if cmd == "zip" {
		err = z.ZipDir(src, dst, true, true)
	} else if cmd == "unzip" {
		err = z.Unzip(src, dst)
	} else if cmd == "zipfiles" {
		dst = os.Args[2]
		err = z.ZipFile(os.Args[3:], os.Args[2], true)
	} else {
		help()
	}
	if err != nil {
		fmt.Println(err)
		_ = os.RemoveAll(dst)
	}
}

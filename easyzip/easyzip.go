/*********************************************************
          File Name: easyzip.go
          Author: Abby Cin
          Mail: abbytsing@gmail.com
          Created Time: Sat 19 Jan 2019 02:31:27 PM CST
**********************************************************/

package easyzip

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

type Zip struct {
	progress func(l string)
}

// create Zip instance, if verbose is true, it will print
// file name when archiving file to zip or extracting file
// from zip.
func NewZip(verbose bool) *Zip {
	res := new(Zip)
	res.initialize(verbose)
	return res
}

func toSlash(p string) string {
	if runtime.GOOS == "windows" {
		return filepath.ToSlash(p)
	}
	return p
}

func absPath(p string) (string, error) {
	r, e := filepath.Abs(p)
	if e != nil {
		return "", e
	}
	return toSlash(r), e
}

func basePath(p string) string {
	return toSlash(filepath.Base(p))
}

func dirPath(p string) string {
	return toSlash(filepath.Dir(p))
}

func (z *Zip) initialize(verbose bool) {
	if verbose {
		z.progress = func(l string) { fmt.Println("add:", l) }
	} else {
		z.progress = func(l string) {}
	}
}

// create a zip file named `dst` from a list of items
// the items can be either file or directory. when error
// occurred, it simply return that error, caller may delete
// `dst` manually.
func (z *Zip) ZipFile(src []string, dst string) error {
	abs_dst, e := absPath(dst)
	if e != nil {
		return e
	}

	var abs_src []string
	for _, s := range src {
		t, e := absPath(s)
		if e != nil {
			return e
		}
		abs_src = append(abs_src, t)
	}
	o, e := os.Create(abs_dst)
	if e != nil {
		return e
	}

	w := zip.NewWriter(o)

	defer func() {
		_ = w.Close()
		_ = o.Close()
	}()

	for _, f := range abs_src {
		e = addFiles(w, f, abs_dst, basePath(f), z.progress)
		if e != nil {
			return e
		}
	}
	return nil
}

// create a zip file from given directory.
// when overwrite enabled, it will delete dst first (if exist), or else return a dst already existed error.
// when create_root enabled, it will create a directory named src and put all contents into it.
func (z *Zip) ZipDir(src, dst string, overwrite, create_root bool) error {
	abs_src, _ := absPath(src)
	abs_dst, _ := absPath(dst)
	i, e := os.Stat(abs_src)
	if e != nil {
		if os.IsNotExist(e) {
			return fmt.Errorf("%s not exist", src)
		} else {
			return e
		}
	}
	if !i.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}
	_, e = os.Stat(abs_dst)
	if e != nil {
		if !os.IsNotExist(e) {
			return e
		}
	}

	if overwrite {
		_ = os.RemoveAll(abs_dst)
	} else {
		return fmt.Errorf("%s exist, skip", dst)
	}

	f, e := os.Create(abs_dst)
	if e != nil {
		return e
	}
	w := zip.NewWriter(f)
	dst = ""
	if create_root {
		dst = basePath(abs_src)
	}
	err := addFiles(w, abs_src, abs_dst, dst, z.progress)
	_ = w.Close()
	_ = f.Close()
	return err
}

func addFiles(z *zip.Writer, src, self, dst string, cb func(l string)) error {
	if src == self {
		cb("skip self")
		return nil
	}
	i, e := os.Stat(src)
	if e != nil {
		return e
	}

	if i.IsDir() {
		items, e := ioutil.ReadDir(src)
		if e != nil {
			return e
		}
		for _, item := range items {
			var tmp string
			if len(dst) == 0 {
				tmp = item.Name()
			} else {
				tmp = dst + "/" + item.Name()
			}
			e = addFiles(z, src + "/" +item.Name(), self, tmp, cb)
			if e != nil {
				return e
			}
		}
	} else {
		reader, e := os.Open(src)
		if e != nil {
			return e
		}
		writer, e := z.Create(dst)
		if e != nil {
			_ = reader.Close()
			return e
		}
		cb(src)
		_, e = io.Copy(writer, reader)
		_ = reader.Close()
		if e != nil {
			return e
		}
	}
	return nil
}

// unzip src to dst, if dst is empty, the contents will extract to current work directory,
// otherwise put into dst directory.
func (z *Zip) Unzip(src, dst string) error {
	abs_src, e := absPath(src)
	if e != nil {
		return e
	}

	abs_dst := dst
	if len(dst) == 0 {
		abs_dst, e = os.Getwd()
	} else {
		abs_dst, e = absPath(dst)
	}
	if e != nil {
		return e
	}

	r, e := zip.OpenReader(abs_src)
	if e != nil {
		return e
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		file := abs_dst + "/" + f.Name
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(file, 0755)
		} else {
			_ = os.MkdirAll(dirPath(file), 0755)
			reader, e := f.Open()
			if e != nil {
				return e
			}
			writer, e := os.Create(file)
			if e != nil {
				_ = reader.Close()
				return e
			}
			z.progress(file)
			_, e = io.Copy(writer, reader)
			_ = writer.Close()
			_ = reader.Close()
			if e != nil {
				return e
			}
		}
	}
	return nil
}

package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/kr/fs"
)

var (
	fileNamesToPack   = []string{"readme", "license"}
	allowedExtensions = []string{"", ".txt", ".md"}
	osList            = []string{"darwin", "freebsd", "linux", "netbsd", "openbsd", "windows"}
	archList          = []string{"386", "386.exe", "amd64", "amd64.exe", "arm"}
)

func main() {
	if len(os.Args) == 0 {
		log.Fatalln("nothing to tar")
	}
	if err := tarball(os.Args[1:]...); err != nil {
		log.Fatal(err)
	}
}

func tarball(filepaths ...string) error {
	for _, fpath := range filepaths {

		// print for users
		fmt.Println(fpath + ".tar.gz")

		// set up the output file
		file, err := os.Create(fpath + ".tar.gz")
		if err != nil {
			return err
		}
		defer file.Close()

		// set up the gzip writer
		gw := gzip.NewWriter(file)
		defer gw.Close()
		tw := tar.NewWriter(gw)
		defer tw.Close()

		filepaths := []string{fpath}
		w := fs.Walk(".")
		w.Step()
		for w.Step() {

			fstat := w.Stat()
			if fstat.IsDir() {
				w.SkipDir()
				continue
			}
			if w.Err() != nil {
				return w.Err()
			}

			fname := fstat.Name()
			ext := path.Ext(fname)
			for _, allowed := range allowedExtensions {
				if ext == allowed {
					for _, fileToPack := range fileNamesToPack {
						fnameWithoutExt := fname[:len(fstat.Name())-len(ext)]
						if strings.ToLower(fnameWithoutExt) == fileToPack {
							filepaths = append(filepaths, fname)
							break
						}
					}
					break
				}
			}
		}

		for _, fpath := range filepaths {
			if err := addFile(tw, fpath); err != nil {
				log.Fatalln(err)
			}
		}
	}
	return nil
}

func addFile(tw *tar.Writer, path string) error {

	// open a file, we will close it manually later
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	// get info on that file
	finfo, err := file.Stat()
	if err != nil {
		return err
	}

	// default the name of the file to its current name
	name := finfo.Name()

	// if the name of the file contains two "_" we can assume it is probably a
	// default gox naming template convention of "{{.Dir}}_{{.OS}}_{{.Arch}}".
	if nameslice := strings.Split(name, "_"); len(nameslice) == 3 {

		// check the os name value is valid
		for _, osname := range osList {
			if nameslice[1] == osname {
				name = nameslice[0]
				break
			}
		}

		// check the arch name value is valid
		for i, arch := range archList {
			if nameslice[2] == arch {
				break
			}
			// if this is the last value in the slice, it means we have not matched
			// any arch values, and maybe not os values, therefore default back to
			// the original file name
			if i == len(osList)-1 {
				name = finfo.Name()
			}
		}

		// if the name has changed, we will need to update file info
		if name != finfo.Name() {

			// first rename the binary to its normal name
			if err := os.Rename(finfo.Name(), name); err != nil {
				return err
			}

			// close the file
			file.Close()

			// open the renamed file
			file, err = os.Open(name)
			if err != nil {
				return err
			}

			// update the file info
			if finfo, err = file.Stat(); err != nil {
				return err
			}
		}

		// now lets create the header as needed for this file within the tarball
		header, err := tar.FileInfoHeader(finfo, finfo.Name())
		if err != nil {
			return err
		}

		// write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// copy the file data to the tarball
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}

		// remove the binary since we may have renamed it
		if err := os.Remove(finfo.Name()); err != nil {
			return err
		}
	}

	// close the file
	file.Close()
	return nil
}

package main

import (
	"archive/tar"
	"compress/gzip"
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

		// grab the paths that need to be added in
		paths := []string{fpath}

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
							paths = append(paths, fname)
							break
						}
					}
					break
				}
			}
		}

		// add each file as needed into the current tar archive
		for i := range paths {
			if err := addFile(tw, paths[i]); err != nil {
				log.Fatalln(err)
			}
		}
	}

	return nil
}

func addFile(tw *tar.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if stat, err := file.Stat(); err == nil {

		// now lets create the header as needed for this file within the tarball
		header, err := tar.FileInfoHeader(stat, stat.Name())
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
	}
	return nil
}

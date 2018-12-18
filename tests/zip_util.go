package tests

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func recreateDir(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

func createDirForFile(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}

func unzipFile(f *zip.File, dstPath string) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	err = createDirForFile(dstPath)
	if err != nil {
		return err
	}

	w, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	if err != nil {
		w.Close()
		os.Remove(dstPath)
		return err
	}
	err = w.Close()
	if err != nil {
		os.Remove(dstPath)
		return err
	}
	return nil
}

// Unzip unzips a given zip file to a given directory
func Unzip(zipPath string, destDir string) error {
	st, err := os.Stat(zipPath)
	if err != nil {
		return err
	}
	fileSize := st.Size()
	f, err := os.Open(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zr, err := zip.NewReader(f, fileSize)
	if err != nil {
		return err
	}
	err = recreateDir(destDir)
	if err != nil {
		return err
	}

	for _, fi := range zr.File {
		if fi.FileInfo().IsDir() {
			continue
		}
		destPath := filepath.Join(destDir, fi.Name)
		err = unzipFile(fi, destPath)
		if err != nil {
			os.RemoveAll(destDir)
			return err
		}
	}
	return nil
}

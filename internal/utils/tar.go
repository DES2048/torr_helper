package utils

import (
	"archive/tar"
	"io"
	"log"
	"os"
)

func CreateTar(tarPath string, path string, arcname string) error {
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()
	tw := tar.NewWriter(tarFile)
	defer tw.Close()
	TraverseDirectory(path, arcname, tw)
	return nil
}

func WriteToTar(path string, arcpath string, tw *tar.Writer, fi os.FileInfo) {
	// Open the path
	fr, _ := os.Open(path)
	defer fr.Close()

	// create new header and update the details accrodingly
	h := new(tar.Header)
	if fi.IsDir() {
		h.Typeflag = tar.TypeDir
	} else {
		h.Typeflag = tar.TypeReg
	}
	h.Name = arcpath // u can modify this based on your requirement
	h.Size = fi.Size()
	h.Mode = int64(fi.Mode())
	h.ModTime = fi.ModTime()
	_ = tw.WriteHeader(h)

	if !fi.IsDir() {
		_, _ = io.Copy(tw, fr)
	}
}

// Move inside each directory and write info to tar
// dirPath : folder which you want to tar it.
// tw      : its tarFile writer to your tar file.
func TraverseDirectory(dirPath string, arcpath string, tw *tar.Writer) {
	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	defer dir.Close()
	// read all the files/dir in it
	fis, err := dir.Readdir(0)
	if err != nil {
		log.Fatal(err)
	}

	for _, fi := range fis {
		curPath := dirPath + "/" + fi.Name()

		WriteToTar(curPath, arcpath+"/"+fi.Name(), tw, fi)
		if fi.IsDir() {
			TraverseDirectory(curPath, arcpath+"/"+fi.Name(), tw)
		}
	}
}

package arch

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

type CompressionRequest struct {
	ArchiveName string `json:"file"`
	Directory   string `json:"dir"`
}

func Compress(req CompressionRequest) error {
	archDir := filepath.Dir(req.ArchiveName)
	os.MkdirAll(archDir, os.ModePerm)

	archiveFile, err := os.Create(req.ArchiveName)
	if err != nil {
		return fmt.Errorf("unable to create archive file (%w)", err)
	}
	defer archiveFile.Close()
	writer := zip.NewWriter(archiveFile)
	defer writer.Close()

	files, err := ioutil.ReadDir(req.Directory)
	if err != nil {
		return fmt.Errorf("unable to list directory %s (%w)", req.Directory, err)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size() > files[j].Size()
	})

	for i := 0; i < 10; i++ {
		filename := files[i].Name()
		fullName := filepath.Join(req.Directory, filename)
		processFile(fullName, writer)
	}

	return nil
}

func processFile(filename string, writer *zip.Writer) error {
	srcFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("unable to open file %s to compress (%w)", filename, err)
	}
	defer srcFile.Close()

	destWr, err := writer.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to add file %s to archive (%w)", filename, err)
	}
	io.Copy(destWr, srcFile)

	return nil
}

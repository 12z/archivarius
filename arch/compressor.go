package arch

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

const defaultMaxFiles = 10

type CompressionRequest struct {
	ArchiveName string `json:"file"`
	Directory   string `json:"dir"`
	Filter      string `json:"filter,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

func Compress(req CompressionRequest) (int, error) {
	archDir := filepath.Dir(req.ArchiveName)
	err := os.MkdirAll(archDir, os.ModePerm)
	if err != nil {
		return http.StatusBadRequest,
			fmt.Errorf("unable to create parent directory for archive (%w)", err)
	}

	archiveFile, err := os.Create(req.ArchiveName)
	if err != nil {
		return http.StatusBadRequest,
			fmt.Errorf("unable to create archive file (%w)", err)
	}
	defer archiveFile.Close()
	writer := zip.NewWriter(archiveFile)
	defer writer.Close()

	// room for optimisation
	files, err := os.ReadDir(req.Directory)
	if err != nil {
		return http.StatusBadRequest,
			fmt.Errorf("unable to list directory %s (%w)", req.Directory, err)
	}
	fileInfos := make([]fs.FileInfo, 0, len(files))
	for _, entry := range files {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return http.StatusInternalServerError,
				fmt.Errorf("error reading info for file %s (%w)", entry.Name(), err)
		}
		fileInfos = append(fileInfos, info)
	}
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].Size() > fileInfos[j].Size()
	})

	maxFiles := defaultMaxFiles
	if req.Limit != 0 {
		maxFiles = req.Limit
	}
	count := 0

	for _, file := range fileInfos {
		filename := file.Name()
		if req.Filter != "" {
			match, err := filepath.Match(req.Filter, filename)
			if err != nil {
				return http.StatusBadRequest, fmt.Errorf("malformed filter (%w)", err)
			}
			if !match {
				continue
			}
		}

		fullName := filepath.Join(req.Directory, filename)
		err := processFile(fullName, writer)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("unable to compress file %s (%w)", filename, err)
		}
		count++
		if count >= maxFiles {
			break
		}
	}

	return http.StatusOK, nil
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

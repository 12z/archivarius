package arch

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type ExtractRequest struct {
	ArchiveName string `json:"file"`
	Directory   string `json:"dir"`
	Filter      string `json:"filter,omitempty"`
}

func Excract(req ExtractRequest) (int, error) {
	reader, err := zip.OpenReader(req.ArchiveName)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("unable to open archive %s (%w)", req.ArchiveName, err)
	}
	defer reader.Close()

	for _, f := range reader.File {
		filename := filepath.Base(f.Name)
		if req.Filter != "" {
			match, err := filepath.Match(req.Filter, filename)
			if err != nil {
				return http.StatusBadRequest, fmt.Errorf("malformed filter (%w)", err)
			}
			if !match {
				continue
			}
		}
		fp := filepath.Join(req.Directory, filename)

		// if folder create it
		if f.FileInfo().IsDir() {
			os.MkdirAll(fp, os.ModePerm)
			continue
		}

		os.MkdirAll(filepath.Dir(fp), os.ModePerm)
		outFile, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return http.StatusBadRequest,
				fmt.Errorf("unable to create file %s (%w)", fp, err)
		}
		defer outFile.Close()

		archFileReader, err := f.Open()
		if err != nil {
			return http.StatusInternalServerError,
				fmt.Errorf("unable to open file %s in archive %s (%w)", fp, req.ArchiveName, err)
		}
		defer archFileReader.Close()

		io.Copy(outFile, archFileReader)
	}

	return http.StatusOK, nil
}

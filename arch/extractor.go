package arch

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Extract(req Request) (int, error) {
	reader, err := zip.OpenReader(req.ArchiveName)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("unable to open archive %s (%w)", req.ArchiveName, err)
	}
	defer reader.Close()

	maxFiles := req.Limit
	count := 0
	// for zip no sorting by size is needed
	// files in reader.File are sorted in order of addition
	// which is by size in our case

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

		if maxFiles == 0 {
			continue
		}
		count++
		if count >= maxFiles {
			break
		}
	}

	return http.StatusOK, nil
}

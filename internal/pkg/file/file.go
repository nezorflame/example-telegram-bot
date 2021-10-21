package file

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// Download downloads the file by its link
func Download(link string) (string, error) {
	resp, err := http.DefaultClient.Get(link)
	if err != nil {
		return "", fmt.Errorf("unable to get file: %w", err)
	}
	defer resp.Body.Close()

	tmpFile, err := NewTemp(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to create temp file: %w", err)
	}
	defer tmpFile.Close()

	return tmpFile.Name(), nil
}

// NewTemp creates a new temporary file
func NewTemp(content io.Reader) (*os.File, error) {
	file, err := os.CreateTemp("", "*")
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %w", err)
	}

	if content != nil {
		if _, err = io.Copy(file, content); err != nil {
			file.Close()
			return nil, fmt.Errorf("unable to write file content: %w", err)
		}
	}
	return file, nil
}

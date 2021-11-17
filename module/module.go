package module

import (
	"os"
	"path"
)

const (
	OUTPUT_DIR = "./output/"
)

func CreateOutputFile(filename string) (*os.File, error) {
	// If exist directory
	dirPath := path.Join(os.Getenv("WORKSPACE"), OUTPUT_DIR)
	if _, err := os.Open(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0644)
	}
	// Create file path
	filePath := path.Join(os.Getenv("WORKSPACE"), OUTPUT_DIR, filename)
	// Create file
	return os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
}

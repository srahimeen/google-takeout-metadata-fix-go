package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// LogFailedFiles writes a list of failed file paths to a log file in the logs directory.
func LogFailedFiles(failedFiles []string) (bool, error) {
	
	// If there are no errors, do nothing
	if len(failedFiles) == 0 {
		return false, nil
	}

	// Join failed files into content
	content := strings.Join(failedFiles, "\n")

	currentDir, err := os.Getwd()
	if err != nil {
		return false, err
	}

	// Create the logs directory if it doesn't exist
	logsDir := filepath.Join(currentDir, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		err = os.Mkdir(logsDir, 0755) // Create with write permissions for the owner
		if err != nil {
			return false, err
		}
	}

	filePath := filepath.Join(logsDir, "failed_files.txt")

	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}

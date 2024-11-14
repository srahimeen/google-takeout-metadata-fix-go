package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/srahimeen/google-takeout-metadata-fix-go/utils"
)

func main() {
	fmt.Println("Welcome to Google Takeout Metadata Fix!")

	// Prompt the user for the directory path
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the directory path: ")
	dirPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	// Validate the directory path

	fmt.Println("The directory you want to use is: " + dirPath)

	// Trim whitespace and newline characters from the input
	dirPath = strings.TrimSpace(dirPath)

	fmt.Println("The formatted directory is: " + dirPath)

	// Validate if the path is a directory
	if !isDir(dirPath) {
		fmt.Println("Invalid directory path:", dirPath)
		return
	}

	// Use a map to keep track of renamed files
	renamedFiles := make(map[string]bool)

	// Use a map to keep track of HEIC.json files, we only want to rename these HEIC files
	renamedHEICJSONFiles := make(map[string]bool)

	// We need to update files in three phases, walking the filepath each time
	// Phase 1: Rename TS.mp4 files and TS.mp4.*.json files to remove the TS
	//			Rename HEIC.*.json files to jpg.*.json files and collect these in renamedHEICJSONFiles
	// Phase 2: Find the HEIC files associated with the json files in renamedHEICJSONFiles, and rename them to jpg
	// Phase 3: All renaming has been completed, so run exiftool to update metadata

	// Phase 1
	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Handle any errors while walking
		}

		if err := utils.RenameTSMP4Files(path, info, renamedFiles); err != nil {
			return fmt.Errorf("error in renameTSMP4Files: %v", err)
		}

		if err := utils.RenameHEICJSONToJPGJSON(path, info, renamedFiles, renamedHEICJSONFiles); err != nil {
			return fmt.Errorf("error in renameHEICJSONFiles: %v", err)
		}

		return nil
	}); err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	} else {
		fmt.Println("All files processed successfully.")
	}

	// Phase 2
	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// We need renameHEICJSONToJPGJSON to run completely first
		// Then we can rename the associated HEIC files to JPG
		if err := utils.RenameHEICToJPG(path, info, renamedFiles, renamedHEICJSONFiles); err != nil {
			return fmt.Errorf("error in renameHEICToJPG: %v", err)
		}

		return nil
	}); err != nil {
		fmt.Printf("Error during second pass: %v\n", err)
	} else {
		fmt.Println("All files processed successfully.")
	}

	// TODO (maybe): Make another function to go through all files again, skipping files processed by exiftoolMetadataFixFileByFile (track this in a map)
	// These files do not have a JSON metadata file (direct download instead of Google Takeout)
	// Set the date for these files based on filename?

	// Phase 3
	if err := utils.ExiftoolMetadataFixFileByFile(dirPath); err != nil {
		fmt.Println("Error executing exiftool:", err)
		return
	} else {
		fmt.Println("Command executed successfully.")
	}

}

// isDir checks if the given path is a directory
func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
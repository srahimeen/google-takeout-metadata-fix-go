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

	// Use a map to keep track of all renamed files so we don't double rename
	renamedFiles := make(map[string]bool)

	// Use a map to keep track of .HEIC.json files, we only want to rename the matching .HEIC files
	// HEIC files which do not have a json file do not need to be renamed
	renamedHEICJSONFiles := make(map[string]bool)

	// Keep track of the JPG files which were renamed to WEBP
	renamedJPGToWEBPFiles := make(map[string]bool)

	// We need to update files in three phases, walking the filepath each time
	// Phase 1: Rename TS.mp4 files and TS.mp4.*.json files to remove the TS
	//			Rename HEIC.*.json files to jpg.*.json files and collect these in renamedHEICJSONFiles
	//			Some jpg files are secretly webp files, we need to rename the file and the associated json file.
	// Phase 2: Find the HEIC files associated with the json files in renamedHEICJSONFiles, and rename them to jpg
	//			Find the JSON files associated with the jpg files renabed by RenameJPGToWEBP, and rename them to webp
	// Phase 3: All renaming has been completed, so run exiftool to update metadata

	// Phase 1
	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if err := utils.RenameTSMP4Files(path, info, renamedFiles); err != nil {
			return fmt.Errorf("error in renameTSMP4Files: %v", err)
		}

		if err := utils.RenameHEICJSONToJPGJSON(path, info, renamedFiles, renamedHEICJSONFiles); err != nil {
			return fmt.Errorf("error in renameHEICJSONFiles: %v", err)
		}

		if err := utils.RenameJPGToWEBP(path, info, renamedFiles, renamedJPGToWEBPFiles); err != nil {
			return fmt.Errorf("error in RenameJPGToWEBP: %v", err)
		}

		return nil
	}); err != nil {
		fmt.Printf("Error in Phase 1: %v\n", err)
	} else {
		fmt.Println("Phase 1 completed successfully!")
	}

	// Phase 2
	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// We need renameHEICJSONToJPGJSON to run completely first
		if err := utils.RenameHEICToJPG(path, info, renamedFiles, renamedHEICJSONFiles); err != nil {
			return fmt.Errorf("error in renameHEICToJPG: %v", err)
		}

		// We need RenameJPGToWEBP to run completely first
		if err := utils.RenameJPGJSONToWEBPJSON(path, info, renamedFiles, renamedJPGToWEBPFiles); err != nil {
			return fmt.Errorf("error in RenameJPGJSONToWEBPJSON: %v", err)
		}

		return nil
	}); err != nil {
		fmt.Printf("Error in Phase 2: %v\n", err)
	} else {
		fmt.Println("Phase 2 completed successfully!")
	}

	// TODO (maybe): Make another function to go through all files again, skipping files processed by exiftoolMetadataFixFileByFile (track this in a map)
	// These files do not have a JSON metadata file (direct download instead of Google Takeout)
	// Set the date for these files based on filename?

	// Phase 3
	if err := utils.ExiftoolMetadataFixFileByFile(dirPath); err != nil {
		fmt.Println("Error executing exiftool:", err)
		return
	} else {
		fmt.Println("Phase 3 completed successfully!")
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
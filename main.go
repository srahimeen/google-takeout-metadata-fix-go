package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"path/filepath"
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

	fmt.Println("The formatted directory you want to use is: " + dirPath)

	// Validate if the path is a directory
	if !isDir(dirPath) {
		fmt.Println("Invalid directory path:", dirPath)
		return
	}

	// Handle TS.mp4, TS.json.mp4, .HEIC, .HEIC.json files

	if err := renameFiles(dirPath); err != nil {
		fmt.Println("Error renaming files:", err)
		return
	}

	// Update the datetime from json to image/video files in the specified directory and its subdirectories
	
	if err := exiftoolMetadataFix(dirPath); err != nil {
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

// Run exiftool to copy dates from JSON to image/video files
func exiftoolMetadataFix(dirPath string) error {
	// Define the exiftool command and its arguments
	cmd := exec.Command("exiftool",
		"-d", "%s",
		"-tagsfromfile", "%d%f.%e.json",
		"-DateTimeOriginal<PhotoTakenTimeTimestamp",
		"-FileCreateDate<PhotoTakenTimeTimestamp",
		"-FileModifyDate<PhotoTakenTimeTimestamp",
		"-overwrite_original",
		"-ext", "mp4",
		"-ext", "jpg",
		"-ext", "heic",
		"-ext", "mov",
		"-ext", "jpeg",
		"-ext", "png",
		"-ext", "gif",
		"-ext", "webp",
		"-r", ".",
		"-progress",
	)

	// Set the working directory to the user-provided directory path
	cmd.Dir = dirPath

	// Run the command and capture the output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Running exiftool command.")

	// Execute the command
	return cmd.Run()
}

func renameFiles(dirPath string) error {

	// Keep track of renamed file paths
    renamedFiles := make(map[string]bool)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {

		// Uncomment this if we need extra debugging
		// fmt.Println("Current file: " +  path)

		// Skip paths that have already been renamed
		if renamedFiles[path] {
			// Uncomment this if we need extra debugging
			// fmt.Println("File " + path + " has already been renamed, skipping iteration!")
			return nil
		}

		if err != nil {
			return err
		}

		// Rename .TS.mp4 files to .mp4
		if strings.HasSuffix(info.Name(), ".TS.mp4") {

			fmt.Println("Found TS.mp4 file:", info.Name())

			newPath := strings.Replace(path, ".TS.mp4", ".mp4", 1)
			fmt.Printf("Renaming %s to %s\n", path, newPath)
			if err := os.Rename(path, newPath); err != nil {
				fmt.Println("Error renaming file:", err)
			}

			// Record the renamed TS.mp4 file
			renamedFiles[path] = true
		}

		// Rename .TS.mp4.json files to .mp4.json
		if strings.HasSuffix(info.Name(), ".TS.mp4.json") {

			fmt.Println("Found TS.mp4.json file:", info.Name())

			newPath := strings.Replace(path, ".TS.mp4.json", ".mp4.json", 1)
			fmt.Printf("Renaming %s to %s\n", path, newPath)
			if err := os.Rename(path, newPath); err != nil {
				fmt.Println("Error renaming file:", err)
			}

			// Record the renamed TS.mp4.json file
			renamedFiles[path] = true
		}

		// If any .HEIC files have .HEIC.json files, then they are not "true" HEIC
		// and are actually JPG files with the HEIC extension. We can rename them
		// to .JPG and update the .HEIC.json file to .jpg.json as well

		// Check if the current file has a .HEIC extension
		if strings.HasSuffix(info.Name(), ".HEIC") {
			// Construct the expected .HEIC.json filename

			heicJsonFilePath := path + ".json"

			// Check if the .HEIC.json file exists
			if _, err := os.Stat(heicJsonFilePath); err == nil {
				// We only consider HEIC files which also have HEIC.json files
				// Ignore any standalone HEIC files
				fmt.Printf("Found HEIC file %s and metadata file %s\n", path, heicJsonFilePath)

				// Construct the new .jpg filename
				jpgFilePath := strings.TrimSuffix(path, ".HEIC") + ".jpg"

				// Rename the .HEIC file to .jpg
				fmt.Printf("Renaming %s to %s\n", path, jpgFilePath)
				if err := os.Rename(path, jpgFilePath); err != nil {
					return fmt.Errorf("Failed to rename %s to %s: %w", path, jpgFilePath, err)
				}

				// Record the renamed .HEIC file
				renamedFiles[path] = true

				// Construct the new .jpg.json filename
				jpgJsonFilePath := strings.TrimSuffix(heicJsonFilePath, ".HEIC.json") + ".jpg.json"

				// Rename the .HEIC.json file to .jpg.json
				fmt.Printf("Renaming %s to %s\n", heicJsonFilePath, jpgJsonFilePath)
				
				if err := os.Rename(heicJsonFilePath, jpgJsonFilePath); err != nil {
					return fmt.Errorf("Failed to rename %s to %s: %w", heicJsonFilePath, jpgJsonFilePath, err)
				}

				// Record the renamed .HEIC.json file
                renamedFiles[heicJsonFilePath] = true
			} else {
				fmt.Println("Error with finding HEIC.json file:", err)
			}
		}

		return nil
	})


	if err != nil {
		return err
	}

	return nil
}

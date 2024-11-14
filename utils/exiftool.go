package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// Google Photos now has file names such as filename.extension.supplemental-meta*.json and other variants
// This function handles these cases by going through each JSON file
// Finding the associated media file, and then performing metadata updates
// Media files which have no matching JSON files are ignored
func ExiftoolMetadataFixFileByFile(dirPath string) error {
	fmt.Println("==== FIXING METADATA USING EXIFTOOL ====")

	// Regex to get string until first extension, ie. the media extension (.jpg, .mp4, etc.)
	filenamePattern := regexp.MustCompile(`^(.*?\.\w+)\.`)

	var successCount, errorCount, skipCount int

	failedFiles := []string{}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fmt.Println("Current filepath: " + path)

		// Check if the file has a .json extension
		if !info.IsDir() && filepath.Ext(info.Name()) == ".json" {
			fmt.Println("===")
			fmt.Println("Found JSON metadata file:", path)

			// Extract the path up to the original file extension
			// This grabs the filename up to the media extension file
			// IMG_20201031_144417.jpg.supplemental-metadata.json -> IMG_20201031_144417.jpg
			match := filenamePattern.FindStringSubmatch(path)

			fmt.Printf("Filenames until first extension: %v\n", match)

			if len(match) > 1 { // Otherwise we did not find the media file
				nonJSONFilePath := match[1] // The second item in array is the best formatted string
				fmt.Println("Media filename with extension found:", nonJSONFilePath)

				// TODO: check if DateTimeOriginal, if does then skip
				// TODO: make another function that uses DateTimeOriginal to update the values that runs before this one

				// Use exiftool to update metadata from JSON to the media file
				cmd := exec.Command("exiftool",
					"-d", "%s",
					"-tagsfromfile", path, // JSON file as the source of metadata
					"-DateTimeOriginal<PhotoTakenTimeTimestamp",
					"-FileCreateDate<PhotoTakenTimeTimestamp",
					"-FileModifyDate<PhotoTakenTimeTimestamp",
					"-overwrite_original",
					"-ext", "mp4", "-ext", "jpg", "-ext", "heic", "-ext", "mov", "-ext", "jpeg", "-ext", "png", "-ext", "gif", "-ext", "webp",
					nonJSONFilePath, // Target image file as the file to update
				)

				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil {
					fmt.Println("Error updating file:", nonJSONFilePath, "with JSON:", path, "-", err)
					errorCount++
					failedFiles = append(failedFiles, info.Name())
				} else {
					fmt.Println("Updated ", nonJSONFilePath, " with metadata from ", path)
					successCount++
				}
			} else {
				fmt.Println("Media file for this JSON file was not found: ", path)
				skipCount++
			}
		} else {
			// fmt.Println(path + " is not a JSON file, skipping!")
		}
		return nil
	})

	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Errors: %d\n", errorCount)
	fmt.Printf("Skipped: %d\n", skipCount)

	// Write files which failed to be updated to a log file
	logWritten, logError := LogFailedFiles(failedFiles)
	if logError != nil {
		fmt.Println("Error writing log file:", err)
	} else if !logWritten {
		fmt.Println("No failed files to log.")
	} else {
		fmt.Println("Failed files log written successfully.")
	}

	return err
}

// NOTE: Deprecated function, kept here for reference reasons
// This does not work with JSON metadata files in the format of filename.extension.supplemental-bla.json
func ExiftoolMetadataFixBulk(dirPath string) error {
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

	fmt.Println("Running exiftool to fix metadata.")

	// Execute the command
	return cmd.Run()
}
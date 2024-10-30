package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

	// Handle TS.mp4, TS.json.mp4, .HEIC, .HEIC.json files 
	// TODO re-enable after testing?
	// if err := renameFiles(dirPath); err != nil {
	// 	fmt.Println("Error renaming files:", err)
	// 	return
	// }

	// Update the datetime from json to image/video files in the specified directory and its subdirectories
	// This needs to be run after any file renaming/updating has taken place
	// TODO renable once testing is done
	// if err := exiftoolMetadataFix(dirPath); err != nil {
	// 	fmt.Println("Error executing exiftool:", err)
	// 	return
	// } else {
	// 	fmt.Println("Command executed successfully.")
	// }

	if err := exiftoolMetadataFixVariant(dirPath); err != nil {
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

	fmt.Println("Running exiftool to fix metadata.")

	// Execute the command
	return cmd.Run()
}


// ====== START EXPERIMENT =====

// Google Photos now has file names such as filename.extension.supplemental-meta.json and other variants
// Temporarily creating another function which handles these wildcard cases.
func exiftoolMetadataFixVariant(dirPath string) error {
	// Compile a regex to match the path up to the original file extension
	filenamePattern := regexp.MustCompile(`^(.*?\.\w+)\.`)

	var successCount, errorCount, skipCount int

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file has a .json extension
		if !info.IsDir() && filepath.Ext(info.Name()) == ".json" {
			fmt.Println("====")
			fmt.Println("Found JSON metadata file:", path)

			// Extract the path up to the original file extension
			match := filenamePattern.FindStringSubmatch(path)
			if len(match) > 1 {
				nonJSONFilePath := match[1]
				fmt.Println("Media filename with extension:", nonJSONFilePath)

				// Execute exiftool command to update properties from JSON to the image file
				cmd := exec.Command("exiftool",
					"-d", "%s",
					"-tagsfromfile", path, // JSON file as the source
					"-DateTimeOriginal<PhotoTakenTimeTimestamp",
					"-FileCreateDate<PhotoTakenTimeTimestamp",
					"-FileModifyDate<PhotoTakenTimeTimestamp",
					"-overwrite_original",
					"-ext", "mp4", "-ext", "jpg", "-ext", "heic", "-ext", "mov", "-ext", "jpeg", "-ext", "png", "-ext", "gif", "-ext", "webp",
					nonJSONFilePath, // Target image file
				)

				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				// Run the command and check for errors
				if err := cmd.Run(); err != nil {
					fmt.Println("Error updating file:", nonJSONFilePath, "with JSON:", path, "-", err)
					errorCount++
				} else {
					fmt.Println("Updated", nonJSONFilePath, "with metadata from", path)
					successCount++
				}
			} else {
				fmt.Println("Non-JSON filename not found for:", path)
				skipCount++
			}

			fmt.Println("====")
		}
		return nil
	})

	fmt.Printf("  Successful updates: %d\n", successCount)
	fmt.Printf("  Errors encountered: %d\n", errorCount)
	fmt.Printf("  Files skipped: %d\n", skipCount)

	return err
}

// ====== END EXPERIMENT =====

func renameFiles(dirPath string) error {
	// Keep track of renamed file paths
	renamedFiles := make(map[string]bool)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip paths that have already been renamed
		if renamedFiles[path] {
			return nil
		}

		if err := renameTSMP4Files(path, info, renamedFiles); err != nil {
			return err
		}

		if err := renameHEICFiles(path, info, renamedFiles); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func renameTSMP4Files(path string, info os.FileInfo, renamedFiles map[string]bool) error {
	// Rename .TS.mp4 files to .mp4
	if strings.HasSuffix(info.Name(), ".TS.mp4") {
		fmt.Println("Found TS.mp4 file:", info.Name())

		newPath := strings.Replace(path, ".TS.mp4", ".mp4", 1)
		fmt.Printf("Renaming %s to %s\n", path, newPath)
		if err := os.Rename(path, newPath); err != nil {
			fmt.Println("Error renaming file:", err)
			return err
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
			return err
		}

		// Record the renamed TS.mp4.json file
		renamedFiles[path] = true
	}

	return nil
}

func renameHEICFiles(path string, info os.FileInfo, renamedFiles map[string]bool) error {
	// Check if the current file has a .HEIC extension
	if strings.HasSuffix(info.Name(), ".HEIC") {
		// Construct the expected .HEIC.json filename
		heicJsonFilePath := path + ".json"

		// Check if the related .HEIC.json file exists
		if _, err := os.Stat(heicJsonFilePath); err == nil {
			// We only consider HEIC files which also have related HEIC.json files
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
}

// TODO: When we download the same folder like Photos from 2024 multiple times a year, there are new HEIC files because from Takeout
// they are HEIC and our script renamed it to JPG in a past iteration. We need to delete any dupe HEIC files at the end of it all.
// So if IMG_1642.jpg exists, delete IMG_1642.HEIC and so on.

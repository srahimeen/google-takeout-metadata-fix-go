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

	// Keep track of renamed file paths
	// renamedFiles := make(map[string]bool)

	// Handle TS.mp4, TS.json.mp4, .HEIC, .HEIC.json files 

	// TODO TS.mp4 still needs to be done

	// Use a map to keep track of renamed files
	renamedFiles := make(map[string]bool)

	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Handle any errors while walking
		}

		// Call both renaming functions in sequence
		if err := renameHEICJSONFiles(path, info, renamedFiles); err != nil {
			return fmt.Errorf("error in renameHEICJSONFiles: %v", err)
		}
		if err := renameHEICFiles(path, info, renamedFiles); err != nil {
			return fmt.Errorf("error in renameHEICFiles: %v", err)
		}
		return nil
	}); err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	} else {
		fmt.Println("All files processed successfully.")
	}

	// Update the datetime from json to image/video files in the specified directory and its subdirectories
	if err := exiftoolMetadataFixFileByFile(dirPath); err != nil {
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
// This does not work with JSON metadata files in the format of filename.extension.supplemental-bla.json
func exiftoolMetadataFixBulk(dirPath string) error {
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

// Google Photos now has file names such as filename.extension.supplemental-meta*.json and other variants
// This function handles these cases by going file by file and doing the update
func exiftoolMetadataFixFileByFile(dirPath string) error {
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

	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Errors: %d\n", errorCount)
	fmt.Printf("Skipped: %d\n", skipCount)

	return err
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

// Rename HEIC.*.json files to jpg.json files
func renameHEICJSONFiles(path string, info os.FileInfo, renamedFiles map[string]bool) error {
	// info.Name() is the filename only, path is the full path to file

	if renamedFiles[path] == true {
		return nil
	}

	// If its a directory, we don't need to process it, we need to look for files
	if info.IsDir() {
		return nil
	}

	// Check if the file is a JSON file
	if strings.HasSuffix(info.Name(), ".json") {
		fmt.Println("===")
		fmt.Printf("info.Name(): %s\n", info.Name())
		fmt.Printf("Found JSON file: %s\n", path)
		
		// Check if the JSON file belongs to a HEIC parent file
		if strings.Contains(info.Name(), "HEIC") {
			fmt.Println("It is a HEIC file too")

			parts := strings.Split(info.Name(), ".")

			fmt.Printf("Parts array %v\n", parts)

			// Case for JSON files which have supplemental something, ie. IMG_2086.HEIC.supplemental-metadata.json
			if len(parts) >= 2 {
				secondToLast := parts[len(parts)-2]

				if (strings.Contains(secondToLast, "supple")) {
					fmt.Printf("Second to last part is supplemental: %s\n", secondToLast)

					parts = append(parts[:len(parts)-2], parts[len(parts)-1:]...)
					
					// Check the second to last element again, it should be HEIC
					secondToLast = parts[len(parts)-2]

					if (strings.Contains(secondToLast, "HEIC")) {
						// Update the HEIC to jpg
						parts[len(parts)-2] = "jpg"
					}

					updatedFilenameString := strings.Join(parts, ".")
					fmt.Printf("updatedFilenameString: %s\n", updatedFilenameString)

					// Rename the file
					newPath := filepath.Join(filepath.Dir(path), updatedFilenameString)
					if err := os.Rename(path, newPath); err != nil {
						return fmt.Errorf("failed to rename file: %w", err)
					}
					
					// Document that we have renamed this file
					renamedFiles[path] = true;

					fmt.Printf("HEIC JSON file renamed from %s to: %s\n", info.Name(), newPath)
				}

				// Case for JSON files which have no supplemental string, ie: IMG_2086.HEIC.json
				if (strings.Contains(secondToLast, "HEIC")) {
					fmt.Printf("Second to last part DOES not have supplemental: %s\n", secondToLast)
					
					// Change the HEIC to jpg in the file name
					parts[len(parts)-2] = "jpg"
					updatedFilenameString := strings.Join(parts, ".")

					// Rename the file
					newPath := filepath.Join(filepath.Dir(path), updatedFilenameString)
					if err := os.Rename(path, newPath); err != nil {
						return fmt.Errorf("failed to rename file: %w", err)
					}

					// Document that we have renamed this file
					renamedFiles[path] = true;

					fmt.Printf("HEIC JSON file renamed from %s to: %s\n", info.Name(), newPath)
				}
			}

		} else {
			fmt.Println("This JSON file is NOT for a HEIC file")
		}
	}
	return nil
}

// Rename HEIC files to jpg files
func renameHEICFiles(path string, info os.FileInfo, renamedFiles map[string]bool) error {
	fmt.Println("===")

	// info.Name() is the filename only, path is the full path to file

	if renamedFiles[path] == true {
		return nil
	}

	// If its a directory, we don't need to process it, we need to look for files
	if info.IsDir() {
		return nil
	}

	// Check if the file is a HEIC file
	if  strings.HasSuffix(info.Name(), ".HEIC") {

		fmt.Printf("HEIC file found: %s\n", info.Name())

		newPath := strings.TrimSuffix(path, ".HEIC") + ".jpg"

		if err := os.Rename(path, newPath); err != nil {
			return fmt.Errorf("failed to rename file: %w", err)
		}
		
	}
	return nil
}

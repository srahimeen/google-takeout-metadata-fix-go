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

	// Use a map to keep track of renamed files
	renamedFiles := make(map[string]bool)

	// Use a map to keep track of HEIC.json files, we only want to rename these HEIC files
	renamedHEICJSONFiles := make(map[string]bool)

	// First filepath walk
	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Handle any errors while walking
		}

		if err := renameTSMP4Files(path, info, renamedFiles); err != nil {
			return fmt.Errorf("error in renameTSMP4Files: %v", err)
		}
		if err := renameHEICJSONToJPGJSON(path, info, renamedFiles, renamedHEICJSONFiles); err != nil {
			return fmt.Errorf("error in renameHEICJSONFiles: %v", err)
		}

		return nil
	}); err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	} else {
		fmt.Println("All files processed successfully.")
	}

	// Second filepath walk
	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// We need renameHEICJSONToJPGJSON to run completely first
		// Then we can rename the associated HEIC files to JPG
		if err := renameHEICToJPG(path, info, renamedFiles, renamedHEICJSONFiles); err != nil {
			return fmt.Errorf("error in renameHEICToJPG: %v", err)
		}

		return nil
	}); err != nil {
		fmt.Printf("Error during second pass: %v\n", err)
	} else {
		fmt.Println("All files processed successfully.")
	}

	// TODO: Have a logger writing to file about all the files skipped and errored

	// Make another function to go through all files again, skipping files processed by exiftoolMetadataFixFileByFile (track this in a map)
	// These files do not have a JSON metadata file (direct download instead of Google Takeout)
	// Set the date for these files based on filename?

	// Third filepath walk
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
// This function handles these cases by going through each JSON file
// Finding the associated media file, and then performing metadata updates
// Media files which have no matching JSON files are ignored 
func exiftoolMetadataFixFileByFile(dirPath string) error {
	fmt.Println("==== FIXING METADATA USING EXIFTOOL ====")

	// Regex to get string until first extension, ie. the media extension (.jpg, .mp4, etc.)
	filenamePattern := regexp.MustCompile(`^(.*?\.\w+)\.`)

	var successCount, errorCount, skipCount int

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

	return err
}

// Renames TS.mp4 and TS.mp4.json and TS.mp4.supplemen-bla.json files
// Removes the TS and supplemen-bla portion
func renameTSMP4Files(path string, info os.FileInfo, renamedFiles map[string]bool) error {

	// We have already processed this file
	if renamedFiles[path] == true {
		return nil
	}

	// If its a directory, we don't need to process it, we need to look for files
	if info.IsDir() {
		return nil
	}

	// Rename .TS.mp4 files to .mp4
	if strings.Contains(path, ".TS.") {

		if strings.Contains(path, "json") { // It is a TS.json or TS.supplementa-metadata.json file

			fmt.Println("Found JSON file with TS exension: " + path)
			
			parts := strings.Split(info.Name(), ".")

			fmt.Printf("Parts array %v\n", parts)

			// 1. If it is filename.TS.mp4.json, then parts will be size 4
			// 2. If it is filename.TS.mp4.supplemental-metadata.json, size will be 5

			secondToLast := parts[len(parts)-2]

			if len(parts) == 4 && secondToLast == "mp4" { // Case 1

				parts = append(parts[:1], parts[2:]...)
				newFilename := strings.Join(parts, ".")
				newPath := filepath.Join(filepath.Dir(path), newFilename)

				// fmt.Println("newPath : " + newPath)

				if err := os.Rename(path, newPath); err != nil {
					return fmt.Errorf("Case 1: Failed to rename file: %w", err)
				}

				// Document that we have renamed this file
				renamedFiles[path] = true;

				fmt.Printf("Renamed from %s to %s\n", path, newPath)

			} else if len(parts) == 5 && secondToLast != "mp4" { // Case 2

				// Remove the "TS" part at index 1
				parts = append(parts[:1], parts[2:]...)

				// Now remove the "supplemental-metadata" part, which is at index 2 after the first removal
				parts = append(parts[:2], parts[3:]...)

				newFilename := strings.Join(parts, ".")
				newPath := filepath.Join(filepath.Dir(path), newFilename)

				// fmt.Println("newPath : " + newPath)

				if err := os.Rename(path, newPath); err != nil {
					return fmt.Errorf("Case 2: Failed to rename file: %w", err)
				}

				// Document that we have renamed this file
				renamedFiles[path] = true;

				fmt.Printf("Renamed from %s to %s\n", path, newPath)
			}
		} else { // It is just a TS.mp4 file

			fmt.Println("Found media file with TS exension: " + path)

			// Rename file to remove the .TS sub extension
			newPath := strings.Replace(path, ".TS.", ".", 1)

			err := os.Rename(path, newPath)
			if err != nil {
				return fmt.Errorf("Failed to rename file %s to %s: %w", path, newPath, err)
			}

			// Document that we have renamed this file
			renamedFiles[path] = true;

			fmt.Printf("Renamed %s to %s\n", path, newPath)
		}
	}

	return nil
}

// Rename HEIC.*.json files to jpg.json files
func renameHEICJSONToJPGJSON(path string, info os.FileInfo, renamedFiles map[string]bool, renamedHEICJSONFiles map[string]bool) error {
	// INFO: info.Name() is the filename only, path is the full path to file

	// We already processed this file
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
		// fmt.Printf("info.Name(): %s\n", info.Name())
		// fmt.Printf("Found JSON file: %s\n", path)
		
		// Check if the JSON file belongs to a HEIC parent file
		if strings.Contains(info.Name(), "HEIC") {
			fmt.Println("Found HEIC JSON file: " + path)

			parts := strings.Split(info.Name(), ".")

			// fmt.Printf("Parts array %v\n", parts)

			// Case 1: IMG_2086.HEIC.supplemental-metadata.json, where second to last element is supplemental-blabla
			// Case 2: IMG_2086.HEIC.json, second to last element is HEIC

			if len(parts) >= 2 {
				secondToLast := parts[len(parts)-2]

				if (strings.Contains(secondToLast, "supple")) { // Case 1
					// fmt.Printf("Second to last part is supplemental: %s\n", secondToLast)

					parts = append(parts[:len(parts)-2], parts[len(parts)-1:]...)
					
					// Check the second to last element again, it should be HEIC
					secondToLast = parts[len(parts)-2]

					if (strings.Contains(secondToLast, "HEIC")) {
						// Update the HEIC to jpg
						parts[len(parts)-2] = "jpg"
					}

					updatedFilenameString := strings.Join(parts, ".")
					// fmt.Printf("updatedFilenameString: %s\n", updatedFilenameString)

					// Rename the file
					newPath := filepath.Join(filepath.Dir(path), updatedFilenameString)
					if err := os.Rename(path, newPath); err != nil {
						return fmt.Errorf("Failed to rename file: %w", err)
					}
					
					// Document that we have renamed this file
					renamedFiles[path] = true;

					// Document the HEIC file name
					parts := strings.SplitN(info.Name(), ".", 2)
					trimmedFilename := parts[0] + "." + strings.Split(parts[1], ".")[0] // Up to first extension, ie. IMG_12323.HEIC
					renamedHEICJSONFiles[trimmedFilename] = true;

					fmt.Printf("HEIC JSON file renamed from %s to: %s\n", info.Name(), newPath)
				} else if (strings.Contains(secondToLast, "HEIC")) { // Case 2
					fmt.Printf("Second to last part DOES NOT have supplemental: %s\n", secondToLast)
					
					// Change the HEIC to jpg in the file name
					parts[len(parts)-2] = "jpg"
					updatedFilenameString := strings.Join(parts, ".")

					// Rename the file
					newPath := filepath.Join(filepath.Dir(path), updatedFilenameString)
					if err := os.Rename(path, newPath); err != nil {
						return fmt.Errorf("Failed to rename file: %w", err)
					}

					// Document that we have renamed this file
					renamedFiles[path] = true;

					// Document the HEIC file name
					parts := strings.SplitN(info.Name(), ".", 2)
					trimmedFilename := parts[0] + "." + strings.Split(parts[1], ".")[0] // Up to first extension, ie. IMG_12323.HEIC
					renamedHEICJSONFiles[trimmedFilename] = true;

					fmt.Printf("HEIC JSON file renamed from %s to: %s\n", info.Name(), newPath)
				}
			}

		} else {
			// fmt.Println("This JSON file is NOT for a HEIC file, skipping!")
		}
	} else {
		// fmt.Println("Not a HEIC JSON file, skipping!")
	}

	return nil
}

// Rename HEIC files to jpg files
func renameHEICToJPG(path string, info os.FileInfo, renamedFiles map[string]bool, renamedHEICJSONFiles map[string]bool) error {
	fmt.Println("===")

	// info.Name() is the filename only, path is the full path to file

	if renamedFiles[path] == true {
		return nil
	}

	//fmt.Printf("Renamed HEICJSON files: %v\n", renamedHEICJSONFiles)
	//fmt.Printf("Current HEIC file name: %v\n", info.Name())

	
	// Check if renamedHEICJSONFiles contains this file name
	// We only want to rename the HEIC files whose associated JSON files were also updated
	if renamedHEICJSONFiles[info.Name()] != true {
		return nil
	}

	// If its a directory, we don't need to process it, we need to look for files
	if info.IsDir() {
		return nil
	}

	// Check if the file is a HEIC file like IMG122323.HEIC
	if  strings.HasSuffix(info.Name(), ".HEIC") {

		fmt.Printf("HEIC file found: %s\n", info.Name())

		newPath := strings.TrimSuffix(path, ".HEIC") + ".jpg"

		if err := os.Rename(path, newPath); err != nil {
			return fmt.Errorf("failed to rename file: %w", err)
		}

		// Document that we have renamed this file
		renamedFiles[path] = true;

		fmt.Printf("HEIC file renamed from %s to: %s\n", info.Name(), newPath)
	} else {
		// fmt.Println("Not a HEIC file, no need to rename, skipping!")
	}
	return nil
}

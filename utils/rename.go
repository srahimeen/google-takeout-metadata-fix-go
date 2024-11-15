package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Renames TS.mp4 and TS.mp4.json and TS.mp4.supplemen-bla.json files
// Removes the TS and supplemen-bla portion
func RenameTSMP4Files(path string, info os.FileInfo, renamedFiles map[string]bool) error {

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
func RenameHEICJSONToJPGJSON(path string, info os.FileInfo, renamedFiles map[string]bool, renamedHEICJSONFiles map[string]bool) error {
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
		// fmt.Printf("info.Name(): %s\n", info.Name())
		// fmt.Printf("Found JSON file: %s\n", path)
		
		// Check if the JSON file belongs to a HEIC parent file
		if strings.Contains(info.Name(), "HEIC") {

			fmt.Println("===")
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
func RenameHEICToJPG(path string, info os.FileInfo, renamedFiles map[string]bool, renamedHEICJSONFiles map[string]bool) error {
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
	if strings.HasSuffix(info.Name(), ".HEIC") {

		fmt.Println("===")
		fmt.Printf("HEIC file found: %s\n", info.Name())

		newPath := strings.TrimSuffix(path, ".HEIC") + ".jpg"

		if err := os.Rename(path, newPath); err != nil {
			return fmt.Errorf("Failed to rename file: %w", err)
		}

		// Document that we have renamed this file
		renamedFiles[path] = true;

		fmt.Printf("HEIC file renamed from %s to: %s\n", info.Name(), newPath)
	} else {
		// fmt.Println("Not a HEIC file, no need to rename, skipping!")
	}
	return nil
}

// Uses exiftool to get the file type of a given file
func checkFileType(filePath string) (string, error) {
	cmd := exec.Command("exiftool", "-FileType", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing exiftool: %w", err)
	}

	// Parse output to extract the file type
	output := strings.TrimSpace(out.String())
	if strings.Contains(output, ": ") {
		return strings.Split(output, ": ")[1], nil
	}
	return "Unknown", nil
}

// Some JPG files are actually WEBP files, this renames them to match
func RenameJPGToWEBP(path string, info os.FileInfo, renamedFiles map[string]bool, renamedJPGToWEBPFiles map[string]bool) error {
	if renamedFiles[path] == true {
		return nil
	}

	// If its a directory, we don't need to process it, we need to look for files
	if info.IsDir() {
		return nil
	}

	// Get the file extension
	fileExtension := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")

	// Check file type using EXIF data
	fileType, err := checkFileType(path)
	if err != nil {
		fmt.Printf("Failed to get file type for %s: %v\n", path, err)
		return nil
	}

	if strings.Contains(fileType, "WEBP") {

		// WEBP has mismatched jpg extension
		if (strings.Contains(strings.ToLower(fileExtension), "jpg")) {
			// fmt.Printf("%s is a WEBP, and the extension is %s\n", path, fileExtension)

			// Rename this file to change the extension to webp
			newPath := strings.TrimSuffix(path, ".jpg") + ".webp"

			if err := os.Rename(path, newPath); err != nil {
				return fmt.Errorf("failed to rename %s to %s: %w", path, newPath, err)
			}

			// Record that we have renamed this JPG file to WEBP
			// We need to update the associated JSON file later

			// Get the base name of the file until the first extension (.)
			nameWithoutExt := strings.SplitN(info.Name(), ".", 2)[0]

			// fmt.Printf("Filename without any extensions: %s\n", nameWithoutExt)

			// Record the renamed file by its name without extensions
			renamedJPGToWEBPFiles[nameWithoutExt] = true

			renamedFiles[path] = true

			fmt.Printf("JPG file renamed to WEBP from %s to: %s\n", info.Name(), newPath)
		}

		return nil
	}
	
	return nil
}

// Update the JSON files of the files updated by RenameJPGToWEBP to match the new name
func RenameJPGJSONToWEBPJSON(path string, info os.FileInfo, renamedFiles map[string]bool, renamedJPGToWEBPFiles map[string]bool) error {
	if renamedFiles[path] {
		return nil
	}

	// If its a directory, we don't need to process it, we need to look for files
	if info.IsDir() {
		return nil
	}

	fmt.Printf("Current file: %s\n", path)


	// Get the base name of the file until the first extension (.)
	nameWithoutExt := strings.SplitN(info.Name(), ".", 2)[0]

	if  renamedJPGToWEBPFiles[nameWithoutExt] && strings.HasSuffix(info.Name(), "json") {
		// This is a JSON file whose JPG file was renamed to WEBP
		// Rename this file to match

		newPath := strings.Replace(path, ".jpg.", ".webp.", 1)

		if err := os.Rename(path, newPath); err != nil {
			return fmt.Errorf("Failed to rename file: %w", err)
		}

		fmt.Printf("Renamed JSON (for JPG->WEBP) file from %s to %s\n", path, newPath)
		
	}

	return nil
}

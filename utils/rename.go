package utils

import (
	"fmt"
	"os"
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
	if  strings.HasSuffix(info.Name(), ".HEIC") {

		fmt.Println("===")
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
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

	// VALIDATE THE DIRECTORY PATH

	fmt.Println("The directory you want to use is: " + dirPath)

	// Trim whitespace and newline characters from the input
	dirPath = strings.TrimSpace(dirPath)

	fmt.Println("The formatted directory you want to use is: " + dirPath)

	// Validate if the path is a directory
	if !isDir(dirPath) {
		fmt.Println("Invalid directory path:", dirPath)
		return
	}

	// TODO: RENAME ALL .TS.mp4 files to just .mp4 AND .TS.mp4.json files to just mp4.json
	if err := renameTSmp4Files(dirPath); err != nil {
		fmt.Println("Error renaming files:", err)
		return
	}

	
	// TODO: CONVERT HEIC FILES TO jpg
	// Do this only if a JSON file for it exists
	// check if the file is HEIC
	// if it is HEIC, see if a JSON file for it exists
	// if it exists then
	// convert the HEIC file to jpg
	// rename the HEIC.json file to jpg.json
	// Otherwise do nothing since the HEIC was manually downloaded and is a true HEIC with proper metadata

	// UPDATE THE DATETIME FROM JSON TO IMAGE AND VIDEO FILES

	// Fix metadata in the specified directory and its subdirectories
	if err := exiftoolMetadataFix(dirPath); err != nil {
		fmt.Println("Error executing exiftool:", err)
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


// Rename .TS.mp4 and .TS.mp4.json files in the directory and subdirectories
func renameTSmp4Files(dirPath string) error {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
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
		}

		// Rename .TS.mp4.json files to .mp4.json
		if strings.HasSuffix(info.Name(), ".TS.mp4.json") {

			fmt.Println("Found TS.mp4.json file:", info.Name())

			newPath := strings.Replace(path, ".TS.mp4.json", ".mp4.json", 1)
			fmt.Printf("Renaming %s to %s\n", path, newPath)
			if err := os.Rename(path, newPath); err != nil {
				fmt.Println("Error renaming file:", err)
			}
		}

		return nil
	})

	return err
}

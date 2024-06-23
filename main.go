package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	apiURL      = "https://raw.githubusercontent.com/amcsi/magyar-master-duel-hex-kodok/master/codes.txt" // Replace with your actual API endpoint
	version     = "1.0.0"
	mainDataDir = "main_data"
)

func main() {
	// Check if the script is running with elevated privileges
	if !isElevated() {
		fmt.Println("Requesting elevated privileges to create symbolic links...")
		requestElevation()
		return
	}

	// Output the version number of the script
	fmt.Printf("Script Version: %s\n", version)

	// Wait for the user to press any key to continue
	fmt.Println("Press ENTER to continue...")
	fmt.Scanln()

	// Step 1: Call the API endpoint and get the plain text response
	response, err := http.Get(apiURL)
	if err != nil {
		log.Fatalf("Error fetching API data: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Error reading API response: %v", err)
	}

	folders := strings.Split(string(body), "\n")
	for i := range folders {
		folders[i] = strings.TrimSpace(folders[i])
	}

	// Step 2: Check for the "main_data" folder
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	mainDataPath := filepath.Join(cwd, mainDataDir)
	if _, err := os.Stat(mainDataPath); os.IsNotExist(err) {
		// Find the first folder and rename it to "main_data"
		files, err := ioutil.ReadDir(cwd)
		if err != nil {
			log.Fatalf("Error reading directory: %v", err)
		}

		folderRenamed := false
		for _, file := range files {
			if file.IsDir() {
				oldPath := filepath.Join(cwd, file.Name())
				err = os.Rename(oldPath, mainDataPath)
				if err != nil {
					log.Fatalf("Error renaming folder: %v", err)
				}
				folderRenamed = true
				break
			}
		}

		if !folderRenamed {
			err = os.Mkdir(mainDataPath, os.ModePerm)
			if err != nil {
				log.Fatalf("Error creating main_data folder: %v", err)
			}
		}
	}

	// Step 3: Create or modify folders based on the strings received
	for _, folder := range folders {
		if folder == "" {
			continue
		}

		folderPath := filepath.Join(cwd, folder)
		linkTarget := mainDataPath

		if info, err := os.Lstat(folderPath); err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				// a) Folder is a symlink, leave it alone
				fmt.Printf("Folder %s is a symlink, leaving it alone.\n", folder)
				continue
			} else {
				// b) Folder is a regular folder, delete it and create a symlink
				err = os.RemoveAll(folderPath)
				if err != nil {
					log.Fatalf("Error removing folder %s: %v", folder, err)
				}
				err = os.Symlink(linkTarget, folderPath)
				if err != nil {
					log.Fatalf("Error creating symlink for folder %s: %v", folder, err)
				}
				fmt.Printf("Folder %s was a regular folder, replaced it with a symlink.\n", folder)
			}
		} else if os.IsNotExist(err) {
			// c) Folder did not exist, create a symlink
			err = os.Symlink(linkTarget, folderPath)
			if err != nil {
				log.Fatalf("Error creating symlink for folder %s: %v", folder, err)
			}
			fmt.Printf("Folder %s did not exist, created a symlink.\n", folder)
		} else {
			log.Fatalf("Error checking folder %s: %v", folder, err)
		}
	}

	// Step 4: Show a confirmation message before closing
	fmt.Println("Operation completed. Press Enter to exit.")
	fmt.Scanln()
}

// isElevated checks if the script is running with elevated privileges
func isElevated() bool {
	c := exec.Command("powershell", "-Command", "([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)")
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := c.Output()
	if err != nil {
		log.Fatalf("Error checking elevation status: %v", err)
	}
	return strings.TrimSpace(string(output)) == "True"
}

// requestElevation relaunches the script with elevated privileges
func requestElevation() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}

	c := exec.Command("powershell", "-Command", "Start-Process", exe, "-Verb", "runAs")
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err = c.Start()
	if err != nil {
		log.Fatalf("Error requesting elevation: %v", err)
	}
	fmt.Println("Elevated privileges requested. Please approve the UAC prompt.")
	os.Exit(0)
}

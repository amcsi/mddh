package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

const apiURL = "http://example.com/api/endpoint" // Replace with your actual API endpoint

func main() {
    // Step 1: Call the API endpoint and get the JSON array of strings
    response, err := http.Get(apiURL)
    if err != nil {
        log.Fatalf("Error fetching API data: %v", err)
    }
    defer response.Body.Close()

    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatalf("Error reading API response: %v", err)
    }

    var folders []string
    err = json.Unmarshal(body, &folders)
    if err != nil {
        log.Fatalf("Error unmarshalling JSON: %v", err)
    }

    // Step 2: Check for the "main_data" folder
    cwd, err := os.Getwd()
    if err != nil {
        log.Fatalf("Error getting current working directory: %v", err)
    }

    mainDataPath := filepath.Join(cwd, "main_data")
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
        folderPath := filepath.Join(cwd, folder)
        linkTarget := filepath.Join(cwd, "main_data")

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
package docker

import (
	"dockyard/pkg/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func LoadProjectsFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &Projects)
	if err != nil {
		return err
	}

	return nil
}

func SaveProjectsToFile(filename string) error {
	data, err := json.Marshal(Projects)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func CheckAndLoadProjectsFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		var createFile string
		prompt := &survey.Select{
			Message: fmt.Sprintf("The project configuration file '%s' does not exist. Would you like to create one and add projects? 😎", filePath),
			Options: []string{"Yes", "No"},
		}
		err := survey.AskOne(prompt, &createFile)
		if err != nil {
			return err
		}
		if createFile == "Yes" {
			err := AddProject()
			if err != nil {
				return err
			}
			err = SaveProjectsToFile(filePath)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("No projects file found. Exiting.")
			os.Exit(1)
		}
	} else {
		if err := LoadProjectsFromFile(filePath); err != nil {
			return fmt.Errorf("failed to load projects: %v", err)
		}
	}
	return nil
}

func BrowseForProjectPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}
	currentDir := homeDir

	for {
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return "", fmt.Errorf("failed to read directory: %v", err)
		}

		var options []string
		var paths []string

		if currentDir != "/" {
			options = append(options, ".. (Go up)")
			paths = append(paths, filepath.Dir(currentDir))
		}

		options = append(options, ". (Select current directory)")
		paths = append(paths, currentDir)

		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				options = append(options, fmt.Sprintf("📁 %s/", entry.Name()))
				paths = append(paths, filepath.Join(currentDir, entry.Name()))
			}
		}

		var selectedOption string
		prompt := &survey.Select{
			Message: fmt.Sprintf("Navigate to select project directory (Current: %s):", currentDir),
			Options: options,
		}

		err = survey.AskOne(prompt, &selectedOption)
		if err != nil {
			return "", err
		}

		selectedIndex := -1
		for i, option := range options {
			if option == selectedOption {
				selectedIndex = i
				break
			}
		}

		if selectedIndex == -1 {
			return "", fmt.Errorf("invalid selection")
		}

		selectedPath := paths[selectedIndex]

		if strings.HasPrefix(selectedOption, ". (Select current directory)") {
			return currentDir, nil
		}

		currentDir = selectedPath
	}
}

// HasDockerFiles checks if the directory contains Docker-related files
// Updated to use the improved Docker Compose file detection
func HasDockerFiles(dirPath string) bool {
	// Check for Docker Compose files using the improved detection
	if utils.HasDockerComposeFiles(dirPath) {
		return true
	}

	// Also check for standalone Dockerfile
	dockerfilePath := filepath.Join(dirPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		return true
	}

	return false
}

// GetDockerFilesInfo returns information about Docker files found in the directory
func GetDockerFilesInfo(dirPath string) string {
	var info []string

	// Check for Compose files
	if composeFiles, err := utils.GetAllComposeFiles(dirPath); err == nil {
		for _, file := range composeFiles {
			filename := filepath.Base(file)
			info = append(info, fmt.Sprintf("📄 %s", filename))
		}
	}

	// Check for Dockerfile
	dockerfilePath := filepath.Join(dirPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		info = append(info, "📄 Dockerfile")
	}

	if len(info) == 0 {
		return "No Docker files found"
	}

	return strings.Join(info, ", ")
}

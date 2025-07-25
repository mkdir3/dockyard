package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func ResolveHomeDir(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return strings.Replace(path, "~", usr.HomeDir, 1), nil
}

// GetComposeFilePath finds the Docker Compose file in the project directory
// Supports all standard Docker Compose file names
func GetComposeFilePath(projectDir string) (string, error) {
	// List of compose file names in order of preference
	composeFiles := []string{
		"compose.yaml",
		"compose.yml",
		"docker-compose.yaml",
		"docker-compose.yml",
	}

	for _, filename := range composeFiles {
		filePath := filepath.Join(projectDir, filename)
		if _, err := os.Stat(filePath); err == nil {
			return filePath, nil
		}
	}

	return "", fmt.Errorf("no docker-compose file found in %s. Looking for: %s",
		projectDir, strings.Join(composeFiles, ", "))
}

// GetAllComposeFiles returns all Docker Compose files found in the directory
func GetAllComposeFiles(projectDir string) ([]string, error) {
	composeFiles := []string{
		"compose.yaml",
		"compose.yml",
		"docker-compose.yaml",
		"docker-compose.yml",
		"docker-compose.override.yml",
		"docker-compose.override.yaml",
	}

	var foundFiles []string
	for _, filename := range composeFiles {
		filePath := filepath.Join(projectDir, filename)
		if _, err := os.Stat(filePath); err == nil {
			foundFiles = append(foundFiles, filePath)
		}
	}

	if len(foundFiles) == 0 {
		return nil, fmt.Errorf("no docker-compose files found in %s", projectDir)
	}

	return foundFiles, nil
}

// HasDockerComposeFiles checks if the directory contains any Docker Compose files
func HasDockerComposeFiles(projectDir string) bool {
	files, err := GetAllComposeFiles(projectDir)
	return err == nil && len(files) > 0
}

package docker

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"sort"
)

var Projects = make(map[string]string)

func init() {
	if err := LoadProjectsFromFile("projects.json"); err != nil {
		Projects = make(map[string]string)
	}
}

func GetSortedProjectNames() []string {
	projectNames := make([]string, 0, len(Projects))
	for projectName := range Projects {
		projectNames = append(projectNames, projectName)
	}
	sort.Strings(projectNames)
	return projectNames
}

func AddProject() error {
	var projectName string
	err := survey.AskOne(&survey.Input{Message: "Enter the project name:"}, &projectName)
	if err != nil {
		return fmt.Errorf("failed to get project name: %v", err)
	}

	// Check if project name already exists
	if _, exists := Projects[projectName]; exists {
		var overwrite string
		overwritePrompt := &survey.Select{
			Message: fmt.Sprintf("Project '%s' already exists. Do you want to overwrite it?", projectName),
			Options: []string{"Yes", "No"},
		}
		err = survey.AskOne(overwritePrompt, &overwrite)
		if err != nil {
			return err
		}
		if overwrite == "No" {
			return fmt.Errorf("project addition cancelled")
		}
	}

	fmt.Println("Browse to select the project directory:")
	projectPath, err := BrowseForProjectPath()
	if err != nil {
		return fmt.Errorf("failed to browse for project path: %v", err)
	}

	// Check for Docker files and show detailed information
	if !HasDockerFiles(projectPath) {
		fmt.Printf("⚠️  Warning: No Docker files found in %s\n", projectPath)
		var proceed string
		proceedPrompt := &survey.Select{
			Message: "Do you want to continue anyway?",
			Options: []string{"Yes", "No"},
		}
		err = survey.AskOne(proceedPrompt, &proceed)
		if err != nil {
			return err
		}
		if proceed == "No" {
			return fmt.Errorf("project addition cancelled")
		}
	} else {
		// Show what Docker files were found
		dockerInfo := GetDockerFilesInfo(projectPath)
		fmt.Printf("✅ Found Docker files: %s\n", dockerInfo)
	}

	var confirm string
	confirmPrompt := &survey.Select{
		Message: fmt.Sprintf("Do you want to add project '%s' with path '%s'?", projectName, projectPath),
		Options: []string{"Yes", "No"},
	}
	err = survey.AskOne(confirmPrompt, &confirm)
	if err != nil {
		return err
	}

	if confirm == "Yes" {
		Projects[projectName] = projectPath
		err := SaveProjectsToFile("projects.json")
		if err != nil {
			return err
		}
		fmt.Printf("✅ Successfully added project '%s'\n", projectName)
	} else {
		fmt.Println("Project addition cancelled.")
	}

	return nil
}

func RemoveProject() error {
	if len(Projects) == 0 {
		fmt.Println("No projects found to remove.")
		return nil
	}

	// Show available projects for removal
	projectNames := GetSortedProjectNames()
	var projectToRemove string

	removePrompt := &survey.Select{
		Message: "Select the project you'd like to remove:",
		Options: projectNames,
	}

	err := survey.AskOne(removePrompt, &projectToRemove)
	if err != nil {
		return err
	}

	// Confirm removal
	var confirm string
	confirmPrompt := &survey.Select{
		Message: fmt.Sprintf("Are you sure you want to remove project '%s'?", projectToRemove),
		Options: []string{"Yes", "No"},
	}
	err = survey.AskOne(confirmPrompt, &confirm)
	if err != nil {
		return err
	}

	if confirm == "Yes" {
		delete(Projects, projectToRemove)

		if err := SaveProjectsToFile("projects.json"); err != nil {
			return fmt.Errorf("failed to save projects after removal: %v", err)
		}

		fmt.Printf("✅ Successfully removed project '%s'\n", projectToRemove)
	} else {
		fmt.Println("Project removal cancelled.")
	}

	return nil
}

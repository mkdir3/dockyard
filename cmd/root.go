package cmd

import (
	"dockyard/pkg/docker"
	"dockyard/pkg/utils"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dockyard",
	Short: "CLI app to manage Dockerized projects",
	Long:  `A CLI app to manage Dockerized projects using Docker Compose. It can start, stop and list running Docker containers.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := docker.CheckAndLoadProjectsFile("projects.json"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Check Docker status before proceeding
		if err := docker.CheckDockerStatus(); err != nil {
			fmt.Printf("❌ Docker check failed: %v\n", err)
			return
		}

		// Your existing project selection workflow
		selectedProjects, err := docker.SelectProjects()
		if err != nil {
			fmt.Printf("Failed to select projects: %v\n", err)
			return
		}

		if len(selectedProjects) == 0 {
			fmt.Println("No projects selected.")
			return
		}

		fmt.Printf("🚀 Starting %d selected project(s)...\n\n", len(selectedProjects))

		successCount := 0
		var failedProjects []string

		for _, projectName := range selectedProjects {
			projectPath, ok := docker.Projects[projectName]
			if !ok {
				fmt.Printf("❌ Unknown project: %s\n", projectName)
				failedProjects = append(failedProjects, projectName)
				continue
			}

			projectDir, err := utils.ResolveHomeDir(projectPath)
			if err != nil {
				fmt.Printf("❌ Failed to resolve home directory for %s: %v\n", projectName, err)
				failedProjects = append(failedProjects, projectName)
				continue
			}

			// Create compose manager for each project
			cm, err := docker.NewComposeManager()
			if err != nil {
				fmt.Printf("❌ Failed to create compose manager for %s: %v\n", projectName, err)
				failedProjects = append(failedProjects, projectName)
				continue
			}

			fmt.Printf("📦 Starting project: %s\n", projectName)
			err = cm.StartProject(projectDir, true, true) // detached=true, removeOrphans=true
			err = cm.Close()

			if err != nil {
				fmt.Printf("❌ Failed to start project %s: %v\n", projectName, err)
				failedProjects = append(failedProjects, projectName)

				// If Docker daemon becomes unavailable during operation, stop trying
				if isDaemonError(err) {
					fmt.Println("\n🛑 Docker daemon issue detected. Stopping further operations.")
					break
				}
				continue
			}

			fmt.Printf("✅ Successfully started project: %s\n\n", projectName)
			successCount++
		}

		// Summary with retry option
		fmt.Printf("📊 Summary: %d/%d projects started successfully\n", successCount, len(selectedProjects))

		if len(failedProjects) > 0 {
			fmt.Printf("❌ Failed projects: %v\n", failedProjects)

			// Offer to retry failed projects
			var retryFailed string
			retryPrompt := &survey.Select{
				Message: "Would you like to retry the failed projects?",
				Options: []string{
					"Yes, retry failed projects",
					"No, I'll fix issues manually",
				},
			}

			err = survey.AskOne(retryPrompt, &retryFailed)
			if err == nil && retryFailed == "Yes, retry failed projects" {
				fmt.Println("\n🔄 Retrying failed projects...")
				retryFailedProjects(failedProjects)
			}
		}

		// Show status of all projects if any succeeded
		if successCount > 0 {
			fmt.Println("\n📈 Current project status:")
			showStatusForProjects(selectedProjects)
		} else if len(failedProjects) > 0 {
			fmt.Println("\n💡 Tip: Run 'dockyard status' to check the current state of your projects")
		}
	},
}

// isDaemonError checks if the error is related to Docker daemon connectivity
func isDaemonError(err error) bool {
	errorStr := err.Error()
	daemonErrors := []string{
		"Docker daemon is not running",
		"Cannot connect to the Docker daemon",
		"connection refused",
		"Docker is not accessible",
	}

	for _, daemonError := range daemonErrors {
		if strings.Contains(errorStr, daemonError) {
			return true
		}
	}
	return false
}

func showStatusForProjects(projectNames []string) {
	for _, projectName := range projectNames {
		projectPath, ok := docker.Projects[projectName]
		if !ok {
			continue
		}

		projectDir, err := utils.ResolveHomeDir(projectPath)
		if err != nil {
			fmt.Printf("❌ %s: Failed to resolve path\n", projectName)
			continue
		}

		cm, err := docker.NewComposeManager()
		if err != nil {
			fmt.Printf("❌ %s: Failed to create compose manager\n", projectName)
			continue
		}

		statuses, err := cm.GetProjectStatus(projectDir)
		err = cm.Close()
		if err != nil {
			fmt.Printf("❌ %s: Failed to close compose manager\n", projectName)
			continue
		}

		if len(statuses) == 0 {
			fmt.Printf("📭 %s: No containers\n", projectName)
		} else {
			runningCount := 0
			for _, status := range statuses {
				if status.State == "running" {
					runningCount++
				}
			}

			if runningCount > 0 {
				fmt.Printf("🟢 %s: %d/%d containers running\n", projectName, runningCount, len(statuses))
			} else {
				fmt.Printf("🔴 %s: %d containers stopped\n", projectName, len(statuses))
			}
		}
	}
}

// retryFailedProjects attempts to retry projects that failed to start
func retryFailedProjects(failedProjects []string) {
	successCount := 0
	var stillFailed []string

	for _, projectName := range failedProjects {
		projectPath, ok := docker.Projects[projectName]
		if !ok {
			fmt.Printf("❌ Unknown project: %s\n", projectName)
			stillFailed = append(stillFailed, projectName)
			continue
		}

		projectDir, err := utils.ResolveHomeDir(projectPath)
		if err != nil {
			fmt.Printf("❌ Failed to resolve home directory for %s: %v\n", projectName, err)
			stillFailed = append(stillFailed, projectName)
			continue
		}

		cm, err := docker.NewComposeManager()
		if err != nil {
			fmt.Printf("❌ Failed to create compose manager for %s: %v\n", projectName, err)
			stillFailed = append(stillFailed, projectName)
			continue
		}

		fmt.Printf("🔄 Retrying project: %s\n", projectName)
		err = cm.StartProject(projectDir, true, true)
		err = cm.Close()
		if err != nil {
			fmt.Printf("❌ Failed to start project %s: %v\n", projectName, err)
			stillFailed = append(stillFailed, projectName)

			// If Docker daemon becomes unavailable during operation, stop trying
			if isDaemonError(err) {
				fmt.Println("\n🛑 Docker daemon issue detected. Stopping further operations.")
				break
			}
			continue
		}

		fmt.Printf("✅ Successfully started project: %s\n", projectName)
		successCount++
	}

	fmt.Printf("\n🎯 Retry Results: %d/%d projects started successfully\n", successCount, len(failedProjects))
	if len(stillFailed) > 0 {
		fmt.Printf("❌ Still failing: %v\n", stillFailed)
		fmt.Println("💡 Tip: Use 'dockyard auth' to set up authentication if needed")
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

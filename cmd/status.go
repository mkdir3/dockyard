package cmd

import (
	"dockyard/pkg/docker"
	"dockyard/pkg/utils"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [project]",
	Short: "Show status of Docker project containers",
	Long:  `Display detailed status information for all containers in a project`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Show status for all projects
			showAllProjectsStatus()
			return
		}

		projectName := args[0]
		projectPath, ok := docker.Projects[projectName]
		if !ok {
			fmt.Printf("Unknown project: %s\n", projectName)
			return
		}

		projectDir, err := utils.ResolveHomeDir(projectPath)
		if err != nil {
			fmt.Printf("Failed to resolve home directory in %s: %v\n", projectPath, err)
			return
		}

		showProjectStatus(projectName, projectDir)
	},
}

func showProjectStatus(projectName, projectDir string) {
	// Check Docker status first
	err := docker.CheckDockerStatus()
	if err != nil {
		fmt.Printf("❌ Docker status check failed: %v\n", err)
		fmt.Printf("📁 Project '%s' location: %s\n", projectName, projectDir)
		return
	}

	cm, err := docker.NewComposeManager()
	if err != nil {
		fmt.Printf("Failed to create compose manager: %v\n", err)
		return
	}
	defer func(cm *docker.ComposeManager) {
		err := cm.Close()
		if err != nil {
			fmt.Printf("Failed to close compose manager: %v\n", err)
		} else {
			fmt.Println("✅ Compose manager connection closed")
		}
	}(cm)

	statuses, err := cm.GetProjectStatus(projectDir)
	if err != nil {
		fmt.Printf("Failed to get status for project %s: %v\n", projectName, err)
		return
	}

	if len(statuses) == 0 {
		fmt.Printf("📭 No containers found for project '%s'\n", projectName)
		fmt.Printf("💡 Tip: Run 'dockyard start %s' to create and start containers\n", projectName)
		return
	}

	fmt.Printf("📊 Status for project '%s':\n", projectName)
	fmt.Printf("%-25s %-12s %-10s %-20s %s\n", "SERVICE", "ID", "STATE", "STATUS", "PORTS")
	fmt.Println(strings.Repeat("-", 85))

	for _, status := range statuses {
		stateEmoji := getStateEmoji(status.State)

		fmt.Printf("%-25s %-12s %s%-9s %-20s %s\n",
			status.Service,
			status.ID,
			stateEmoji,
			status.State,
			status.Status,
			status.Ports)
	}
}

func showAllProjectsStatus() {
	fmt.Println("📊 Status for all projects:")
	fmt.Println()

	// Check Docker status first
	err := docker.CheckDockerStatus()
	if err != nil {
		fmt.Printf("❌ Docker status check failed: %v\n", err)
		fmt.Println("📋 Showing project list without container status:")
		fmt.Println()

		// Show projects without Docker status
		sortedProjectNames := docker.GetSortedProjectNames()
		for _, projectName := range sortedProjectNames {
			projectPath := docker.Projects[projectName]
			fmt.Printf("📁 %s: %s\n", projectName, projectPath)
		}
		return
	}

	sortedProjectNames := docker.GetSortedProjectNames()
	for _, projectName := range sortedProjectNames {
		projectPath := docker.Projects[projectName]
		projectDir, err := utils.ResolveHomeDir(projectPath)
		if err != nil {
			fmt.Printf("❌ %s: Failed to resolve path: %v\n", projectName, err)
			continue
		}

		cm, err := docker.NewComposeManager()
		if err != nil {
			fmt.Printf("❌ %s: Failed to create compose manager: %v\n", projectName, err)
			continue
		}

		statuses, err := cm.GetProjectStatus(projectDir)
		cm.Close()

		if err != nil {
			fmt.Printf("❌ %s: Failed to get status: %v\n", projectName, err)
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

			statusEmoji := "⏹️"
			if runningCount > 0 {
				statusEmoji = "🟢"
			}

			fmt.Printf("%s %s: %d/%d containers running\n",
				statusEmoji, projectName, runningCount, len(statuses))
		}
	}
}

func getStateEmoji(state string) string {
	switch state {
	case "running":
		return "🟢 "
	case "exited":
		return "🔴 "
	case "paused":
		return "⏸️  "
	case "restarting":
		return "🔄 "
	default:
		return "⚪ "
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

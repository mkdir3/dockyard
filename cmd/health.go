package cmd

import (
	"dockyard/pkg/docker"
	"dockyard/pkg/utils"
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health [project]",
	Short: "Check and fix project health issues",
	Long:  `Analyze project container health and offer solutions for common issues like stopped containers.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			checkAllProjectsHealth()
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

		checkProjectHealth(projectName, projectDir)
	},
}

func checkAllProjectsHealth() {
	fmt.Println("🏥 Health Check for All Projects")
	fmt.Println("===============================")
	fmt.Println()

	// Check Docker status first
	if err := docker.CheckDockerStatus(); err != nil {
		fmt.Printf("❌ Docker status check failed: %v\n", err)
		return
	}

	sortedProjectNames := docker.GetSortedProjectNames()
	healthyProjects := 0
	var unhealthyProjects []string

	for _, projectName := range sortedProjectNames {
		projectPath := docker.Projects[projectName]
		projectDir, err := utils.ResolveHomeDir(projectPath)
		if err != nil {
			fmt.Printf("❌ %s: Failed to resolve path\n", projectName)
			unhealthyProjects = append(unhealthyProjects, projectName)
			continue
		}

		isHealthy := checkProjectHealthQuiet(projectName, projectDir)
		if isHealthy {
			fmt.Printf("✅ %s: Healthy\n", projectName)
			healthyProjects++
		} else {
			fmt.Printf("⚠️  %s: Needs attention\n", projectName)
			unhealthyProjects = append(unhealthyProjects, projectName)
		}
	}

	fmt.Printf("\n📊 Health Summary: %d healthy, %d need attention\n",
		healthyProjects, len(unhealthyProjects))

	if len(unhealthyProjects) > 0 {
		fmt.Printf("🔧 Projects needing attention: %v\n", unhealthyProjects)

		var fixIssues string
		fixPrompt := &survey.Select{
			Message: "Would you like to fix issues automatically?",
			Options: []string{"Yes, fix all issues", "Let me choose specific projects", "No, I'll handle manually"},
		}

		err := survey.AskOne(fixPrompt, &fixIssues)
		if err != nil {
			return
		}

		switch fixIssues {
		case "Yes, fix all issues":
			fixAllProjectIssues(unhealthyProjects)
		case "Let me choose specific projects":
			selectAndFixProjects(unhealthyProjects)
		}
	}
}

func checkProjectHealth(projectName, projectDir string) {
	fmt.Printf("🏥 Health Check for Project: %s\n", projectName)
	fmt.Println("================================")
	fmt.Println()

	cm, err := docker.NewComposeManager()
	if err != nil {
		fmt.Printf("❌ Failed to create compose manager: %v\n", err)
		return
	}
	defer cm.Close()

	statuses, err := cm.GetProjectStatus(projectDir)
	if err != nil {
		fmt.Printf("❌ Failed to get project status: %v\n", err)
		return
	}

	if len(statuses) == 0 {
		fmt.Printf("📭 No containers found for project '%s'\n", projectName)
		fmt.Printf("💡 Recommendation: Run 'dockyard start %s' to create containers\n", projectName)
		return
	}

	// Analyze container health
	runningCount := 0
	stoppedCount := 0
	errorCount := 0
	var issues []string

	for _, status := range statuses {
		switch status.State {
		case "running":
			runningCount++
		case "exited":
			stoppedCount++
			if strings.Contains(status.Status, "Exited (1)") ||
				strings.Contains(status.Status, "Exited (125)") ||
				strings.Contains(status.Status, "Exited (127)") {
				errorCount++
				issues = append(issues, fmt.Sprintf("❌ %s: %s", status.Service, status.Status))
			} else {
				issues = append(issues, fmt.Sprintf("⏹️  %s: %s", status.Service, status.Status))
			}
		default:
			issues = append(issues, fmt.Sprintf("⚪ %s: %s (%s)", status.Service, status.State, status.Status))
		}
	}

	// Report health status
	if runningCount == len(statuses) {
		fmt.Println("✅ Project is healthy - all containers are running!")
		return
	}

	fmt.Printf("📊 Container Status: %d running, %d stopped (%d with errors)\n",
		runningCount, stoppedCount, errorCount)
	fmt.Println()

	if len(issues) > 0 {
		fmt.Println("🔍 Issues found:")
		for _, issue := range issues {
			fmt.Printf("   %s\n", issue)
		}
		fmt.Println()
	}

	// Offer solutions
	offerHealthSolutions(projectName, projectDir, errorCount > 0, stoppedCount > 0)
}

func checkProjectHealthQuiet(projectName, projectDir string) bool {
	cm, err := docker.NewComposeManager()
	if err != nil {
		return false
	}
	defer cm.Close()

	statuses, err := cm.GetProjectStatus(projectDir)
	if err != nil {
		return false
	}

	if len(statuses) == 0 {
		return false
	}

	// Check if all containers are running
	for _, status := range statuses {
		if status.State != "running" {
			return false
		}
	}

	return true
}

func offerHealthSolutions(projectName, projectDir string, hasErrors, hasStopped bool) {
	var solutions []string

	if hasErrors {
		solutions = append(solutions, "View logs to diagnose errors")
		solutions = append(solutions, "Restart containers with errors")
	}

	if hasStopped {
		solutions = append(solutions, "Start stopped containers")
	}

	solutions = append(solutions, "Full project restart")
	solutions = append(solutions, "Do nothing for now")

	var solution string
	solutionPrompt := &survey.Select{
		Message: "How would you like to fix these issues?",
		Options: solutions,
	}

	err := survey.AskOne(solutionPrompt, &solution)
	if err != nil {
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
			fmt.Printf("❌ Failed to close compose manager: %v\n", err)
		} else {
			fmt.Println("✅ Compose manager connection closed")
		}
	}(cm)

	switch solution {
	case "View logs to diagnose errors":
		fmt.Printf("📋 Viewing logs for project %s:\n", projectName)
		err := cm.ViewLogs(projectDir, []string{}, false)
		if err != nil {
			return
		}

	case "Restart containers with errors", "Start stopped containers", "Full project restart":
		fmt.Printf("🔄 Restarting project %s...\n", projectName)
		err := cm.RestartProject(projectDir)
		if err != nil {
			fmt.Printf("❌ Failed to restart project: %v\n", err)
		} else {
			fmt.Printf("✅ Project %s restarted successfully!\n", projectName)
			fmt.Println("⏳ Checking health in 3 seconds...")

			// Brief pause to let containers start
			time.Sleep(3 * time.Second)

			if checkProjectHealthQuiet(projectName, projectDir) {
				fmt.Println("✅ Project is now healthy!")
			} else {
				fmt.Println("⚠️  Some issues may remain - run health check again if needed")
			}
		}

	case "Do nothing for now":
		fmt.Println("👍 No action taken. You can run this health check again anytime.")
	}
}

func fixAllProjectIssues(projects []string) {
	fmt.Printf("🔧 Fixing issues for %d projects...\n", len(projects))

	for _, projectName := range projects {
		projectPath, ok := docker.Projects[projectName]
		if !ok {
			continue
		}

		projectDir, err := utils.ResolveHomeDir(projectPath)
		if err != nil {
			continue
		}

		fmt.Printf("🔄 Fixing %s...\n", projectName)

		cm, err := docker.NewComposeManager()
		if err != nil {
			fmt.Printf("❌ Failed to fix %s: %v\n", projectName, err)
			continue
		}

		err = cm.RestartProject(projectDir)
		err = cm.Close()
		if err != nil {
			fmt.Printf("❌ Failed to close compose manager for %s: %v\n", projectName, err)
			continue
		} else {
			fmt.Println("✅ Compose manager connection closed")
		}
		fmt.Printf("✅ Fixed %s\n", projectName)
	}
}

func selectAndFixProjects(projects []string) {
	var selectedProjects []string
	prompt := &survey.MultiSelect{
		Message: "Select projects to fix:",
		Options: projects,
	}

	err := survey.AskOne(prompt, &selectedProjects)
	if err != nil {
		return
	}

	if len(selectedProjects) > 0 {
		fixAllProjectIssues(selectedProjects)
	}
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

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

// projectRunner handles the execution of project operations with proper resource management
type projectRunner struct {
	successCount   int
	failedProjects []string
}

// result represents the result of a project operation
type result struct {
	projectName string
	success     bool
	err         error
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "dockyard",
	Short:            "CLI app to manage Dockerized projects",
	Long:             `A CLI app to manage Dockerized projects using Docker Compose. It can start, stop and list running Docker containers.`,
	PersistentPreRun: handlePersistentPreRun,
	PreRun:           handlePreRun,
	Run:              handleRootCommand,
}

// handlePersistentPreRun loads the projects configuration file
func handlePersistentPreRun(cmd *cobra.Command, args []string) {
	if err := docker.CheckAndLoadProjectsFile("projects.json"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// handlePreRun displays project information
func handlePreRun(cmd *cobra.Command, args []string) {
	utils.ProjectInfo()
}

// handleRootCommand is the main entry point for the root command
func handleRootCommand(cmd *cobra.Command, args []string) {
	if err := docker.CheckDockerStatus(); err != nil {
		return
	}

	selectedProjects, err := docker.SelectProjects()
	if err != nil {
		fmt.Printf("Failed to select projects: %v\n", err)
		return
	}

	if len(selectedProjects) == 0 {
		fmt.Println("No projects selected.")
		return
	}

	runner := &projectRunner{}
	runner.startProjects(selectedProjects)
	runner.handleResults(selectedProjects)
}

// startProjects attempts to start all selected projects
func (r *projectRunner) startProjects(selectedProjects []string) {
	fmt.Printf("🚀 Starting %d selected project(s)...\n\n", len(selectedProjects))

	for _, projectName := range selectedProjects {
		result := r.startSingleProject(projectName)
		if result.success {
			fmt.Printf("✅ Successfully started project: %s\n\n", projectName)
			r.successCount++
		} else {
			fmt.Printf("❌ Failed to start project %s: %v\n", projectName, result.err)
			r.failedProjects = append(r.failedProjects, projectName)

			// Stop if Docker daemon becomes unavailable
			if isDaemonError(result.err) {
				fmt.Println("\n🛑 Docker daemon issue detected. Stopping further operations.")
				break
			}
		}
	}
}

// startSingleProject starts a single project and returns the result
func (r *projectRunner) startSingleProject(projectName string) result {
	projectPath, ok := docker.Projects[projectName]
	if !ok {
		return result{
			projectName: projectName,
			success:     false,
			err:         fmt.Errorf("unknown project: %s", projectName),
		}
	}

	projectDir, err := utils.ResolveHomeDir(projectPath)
	if err != nil {
		return result{
			projectName: projectName,
			success:     false,
			err:         fmt.Errorf("failed to resolve home directory: %w", err),
		}
	}

	fmt.Printf("📦 Starting project: %s\n", projectName)
	err = executeWithComposeManager(projectDir, func(cm *docker.ComposeManager) error {
		return cm.StartProject(projectDir, true, true) // detached=true, removeOrphans=true
	})

	return result{
		projectName: projectName,
		success:     err == nil,
		err:         err,
	}
}

// handleResults processes the results and offers retry options
func (r *projectRunner) handleResults(selectedProjects []string) {
	fmt.Printf("📊 Summary: %d/%d projects started successfully\n", r.successCount, len(selectedProjects))

	if len(r.failedProjects) > 0 {
		fmt.Printf("❌ Failed projects: %v\n", r.failedProjects)
		r.offerRetry()
	}

	r.showFinalStatus(selectedProjects)
}

// offerRetry asks the user if they want to retry failed projects
func (r *projectRunner) offerRetry() {
	var retryFailed string
	retryPrompt := &survey.Select{
		Message: "Would you like to retry the failed projects?",
		Options: []string{
			"Yes, retry failed projects",
			"No, I'll fix issues manually",
		},
	}

	if err := survey.AskOne(retryPrompt, &retryFailed); err == nil && retryFailed == "Yes, retry failed projects" {
		fmt.Println("\n🔄 Retrying failed projects...")
		retryFailedProjects(r.failedProjects)
	}
}

// showFinalStatus displays the final status or helpful tips
func (r *projectRunner) showFinalStatus(selectedProjects []string) {
	if r.successCount > 0 {
		fmt.Println("\n📈 Current project status:")
		showStatusForProjects(selectedProjects)
	} else if len(r.failedProjects) > 0 {
		fmt.Println("\n💡 Tip: Run 'dockyard status' to check the current state of your projects")
	}
}

// executeWithComposeManager creates a compose manager, executes the function, and ensures proper cleanup
func executeWithComposeManager(projectDir string, fn func(*docker.ComposeManager) error) error {
	cm, err := docker.NewComposeManager()
	if err != nil {
		return fmt.Errorf("failed to create compose manager: %w", err)
	}
	defer func() {
		if closeErr := cm.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close compose manager: %v\n", closeErr)
		}
	}()

	return fn(cm)
}

// isDaemonError checks if the error is related to Docker daemon connectivity
func isDaemonError(err error) bool {
	if err == nil {
		return false
	}

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

// showStatusForProjects displays the status of all specified projects
func showStatusForProjects(projectNames []string) {
	for _, projectName := range projectNames {
		showSingleProjectStatus(projectName)
	}
}

// showSingleProjectStatus displays the status of a single project
func showSingleProjectStatus(projectName string) {
	projectPath, ok := docker.Projects[projectName]
	if !ok {
		return
	}

	projectDir, err := utils.ResolveHomeDir(projectPath)
	if err != nil {
		fmt.Printf("❌ %s: Failed to resolve path\n", projectName)
		return
	}

	var statuses []docker.ContainerStatus
	err = executeWithComposeManager(projectDir, func(cm *docker.ComposeManager) error {
		var statusErr error
		statuses, statusErr = cm.GetProjectStatus(projectDir)
		return statusErr
	})

	if err != nil {
		fmt.Printf("❌ %s: Failed to get status: %v\n", projectName, err)
		return
	}

	displayProjectStatus(projectName, statuses)
}

// displayProjectStatus formats and displays the container status information
func displayProjectStatus(projectName string, statuses []docker.ContainerStatus) {
	if len(statuses) == 0 {
		fmt.Printf("📭 %s: No containers\n", projectName)
		return
	}

	runningCount := countRunningContainers(statuses)
	if runningCount > 0 {
		fmt.Printf("🟢 %s: %d/%d containers running\n", projectName, runningCount, len(statuses))
	} else {
		fmt.Printf("🔴 %s: %d containers stopped\n", projectName, len(statuses))
	}
}

// countRunningContainers returns the number of running containers
func countRunningContainers(statuses []docker.ContainerStatus) int {
	count := 0
	for _, status := range statuses {
		if status.State == "running" {
			count++
		}
	}
	return count
}

// retryFailedProjects attempts to retry projects that failed to start
func retryFailedProjects(failedProjects []string) {
	retryRunner := &projectRunner{}

	for _, projectName := range failedProjects {
		result := retryRunner.retrySingleProject(projectName)
		if result.success {
			fmt.Printf("✅ Successfully started project: %s\n", projectName)
			retryRunner.successCount++
		} else {
			fmt.Printf("❌ Failed to start project %s: %v\n", projectName, result.err)
			retryRunner.failedProjects = append(retryRunner.failedProjects, projectName)

			// Stop if Docker daemon becomes unavailable
			if isDaemonError(result.err) {
				fmt.Println("\n🛑 Docker daemon issue detected. Stopping further operations.")
				break
			}
		}
	}

	printRetryResults(retryRunner.successCount, len(failedProjects), retryRunner.failedProjects)
}

// retrySingleProject retries starting a single project
func (r *projectRunner) retrySingleProject(projectName string) result {
	fmt.Printf("🔄 Retrying project: %s\n", projectName)
	return r.startSingleProject(projectName)
}

// printRetryResults displays the results of the retry operation
func printRetryResults(successCount, totalRetried int, stillFailed []string) {
	fmt.Printf("\n🎯 Retry Results: %d/%d projects started successfully\n", successCount, totalRetried)
	if len(stillFailed) > 0 {
		fmt.Printf("❌ Still failing: %v\n", stillFailed)
		fmt.Println("💡 Tip: Use 'dockyard auth' to set up authentication if needed")
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

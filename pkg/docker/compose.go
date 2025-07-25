package docker

import (
	"context"
	"dockyard/pkg/utils"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type ComposeManager struct {
	dockerClient client.APIClient
	ctx          context.Context
}

func NewComposeManager() (*ComposeManager, error) {
	ctx := context.Background()

	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}

	return &ComposeManager{
		dockerClient: dockerClient,
		ctx:          ctx,
	}, nil
}

func (cm *ComposeManager) Close() error {
	if cm.dockerClient != nil {
		return cm.dockerClient.Close()
	}
	return nil
}

// ensureDockerRunning checks if Docker is running before executing commands
func (cm *ComposeManager) ensureDockerRunning() error {
	dhc, err := NewDockerHealthChecker()
	if err != nil {
		return fmt.Errorf("failed to create Docker health checker: %v", err)
	}
	defer dhc.Close()

	return dhc.CheckDockerDaemon()
}

// LoadProject loads a Docker Compose project from the project directory
func (cm *ComposeManager) LoadProject(projectDir string) (*types.Project, error) {
	composeFilePath, err := utils.GetComposeFilePath(projectDir)
	if err != nil {
		return nil, err
	}

	// Get project name from directory
	projectName := strings.ToLower(filepath.Base(projectDir))

	// Read the compose file
	composeContent, err := os.ReadFile(composeFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %v", err)
	}

	// Configure loader
	configDetails := types.ConfigDetails{
		WorkingDir: projectDir,
		ConfigFiles: []types.ConfigFile{
			{
				Filename: composeFilePath,
				Content:  composeContent,
			},
		},
		Environment: make(map[string]string),
	}

	// Load environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			configDetails.Environment[parts[0]] = parts[1]
		}
	}

	// Load project with options
	project, err := loader.LoadWithContext(cm.ctx, configDetails, func(options *loader.Options) {
		options.SetProjectName(projectName, true)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load compose project: %v", err)
	}

	return project, nil
}

// GetProjectContainers returns containers for a specific project
func (cm *ComposeManager) GetProjectContainers(projectName string) ([]dockertypes.Container, error) {
	// Check Docker health first
	if err := cm.ensureDockerRunning(); err != nil {
		return nil, fmt.Errorf("docker is not accessible: %v", err)
	}

	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))

	containers, err := cm.dockerClient.ContainerList(cm.ctx, dockertypes.ContainerListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	return containers, nil
}

// StartProject starts all services in the project using docker-compose command
func (cm *ComposeManager) StartProject(projectDir string, detached bool, removeOrphans bool) error {
	// Check Docker health first
	if err := CheckDockerStatus(); err != nil {
		return err
	}

	project, err := cm.LoadProject(projectDir)
	if err != nil {
		return err
	}

	fmt.Printf("🚀 Starting project: %s\n", project.Name)

	// Build docker-compose command
	args := []string{"compose", "-f"}

	composeFilePath, err := utils.GetComposeFilePath(projectDir)
	if err != nil {
		return err
	}
	args = append(args, composeFilePath)

	args = append(args, "up")

	if detached {
		args = append(args, "-d")
	}
	if removeOrphans {
		args = append(args, "--remove-orphans")
	}

	return cm.executeCommandWithErrorHandling(projectDir, args...)
}

// StopProject stops all services in the project
func (cm *ComposeManager) StopProject(projectDir string, removeVolumes bool, removeImages bool) error {
	// Check Docker health first
	if err := CheckDockerStatus(); err != nil {
		return err
	}

	project, err := cm.LoadProject(projectDir)
	if err != nil {
		return err
	}

	fmt.Printf("⏹️  Stopping project: %s\n", project.Name)

	args := []string{"compose", "-f"}

	composeFilePath, err := utils.GetComposeFilePath(projectDir)
	if err != nil {
		return err
	}
	args = append(args, composeFilePath)

	args = append(args, "down")

	if removeVolumes {
		args = append(args, "-v")
	}
	if removeImages {
		args = append(args, "--rmi", "local")
	}

	if err := cm.executeCommandWithErrorHandling(projectDir, args...); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully stopped project: %s\n", project.Name)
	return nil
}

// RestartProject restarts all services in the project
func (cm *ComposeManager) RestartProject(projectDir string) error {
	// Check Docker health first
	if err := CheckDockerStatus(); err != nil {
		return err
	}

	project, err := cm.LoadProject(projectDir)
	if err != nil {
		return err
	}

	fmt.Printf("🔄 Restarting project: %s\n", project.Name)

	composeFilePath, err := utils.GetComposeFilePath(projectDir)
	if err != nil {
		return err
	}

	if err := cm.executeCommandWithErrorHandling(projectDir, "compose", "-f", composeFilePath, "restart"); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully restarted project: %s\n", project.Name)
	return nil
}

// PauseProject pauses all services in the project
func (cm *ComposeManager) PauseProject(projectDir string) error {
	// Check Docker health first
	if err := CheckDockerStatus(); err != nil {
		return err
	}

	project, err := cm.LoadProject(projectDir)
	if err != nil {
		return err
	}

	fmt.Printf("⏸️  Pausing project: %s\n", project.Name)

	composeFilePath, err := utils.GetComposeFilePath(projectDir)
	if err != nil {
		return err
	}

	if err := cm.executeCommandWithErrorHandling(projectDir, "compose", "-f", composeFilePath, "pause"); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully paused project: %s\n", project.Name)
	return nil
}

// UnpauseProject unpauses all services in the project
func (cm *ComposeManager) UnpauseProject(projectDir string) error {
	// Check Docker health first
	if err := CheckDockerStatus(); err != nil {
		return err
	}

	project, err := cm.LoadProject(projectDir)
	if err != nil {
		return err
	}

	fmt.Printf("▶️  Unpausing project: %s\n", project.Name)

	composeFilePath, err := utils.GetComposeFilePath(projectDir)
	if err != nil {
		return err
	}

	if err := cm.executeCommandWithErrorHandling(projectDir, "compose", "-f", composeFilePath, "unpause"); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully unpaused project: %s\n", project.Name)
	return nil
}

// ContainerStatus represents container status information
type ContainerStatus struct {
	Name    string
	Service string
	ID      string
	State   string
	Status  string
	Image   string
	Ports   string
}

// GetProjectStatus returns the status of all containers in the project
func (cm *ComposeManager) GetProjectStatus(projectDir string) ([]ContainerStatus, error) {
	project, err := cm.LoadProject(projectDir)
	if err != nil {
		return nil, err
	}

	containers, err := cm.GetProjectContainers(project.Name)
	if err != nil {
		// If Docker is not accessible, return empty status rather than failing
		if strings.Contains(err.Error(), "Docker is not accessible") {
			return []ContainerStatus{}, nil
		}
		return nil, err
	}

	var statuses []ContainerStatus
	for _, cont := range containers {
		// Extract service name from labels
		serviceName := "unknown"
		if service, ok := cont.Labels["com.docker.compose.service"]; ok {
			serviceName = service
		}

		status := ContainerStatus{
			Name:    strings.TrimPrefix(cont.Names[0], "/"),
			Service: serviceName,
			ID:      cont.ID[:12],
			State:   cont.State,
			Status:  cont.Status,
			Image:   cont.Image,
			Ports:   cm.formatPorts(cont.Ports),
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// ViewLogs displays logs for the project
func (cm *ComposeManager) ViewLogs(projectDir string, services []string, follow bool) error {
	// Check Docker health first
	if err := CheckDockerStatus(); err != nil {
		return err
	}

	composeFilePath, err := utils.GetComposeFilePath(projectDir)
	if err != nil {
		return err
	}

	args := []string{"compose", "-f", composeFilePath, "logs"}

	if follow {
		args = append(args, "-f")
	}

	// Add specific services if provided
	args = append(args, services...)

	return cm.executeCommandWithErrorHandling(projectDir, args...)
}

// PullImages pulls all images for the project
func (cm *ComposeManager) PullImages(projectDir string) error {
	// Check Docker health first
	if err := CheckDockerStatus(); err != nil {
		return err
	}

	project, err := cm.LoadProject(projectDir)
	if err != nil {
		return err
	}

	fmt.Printf("📥 Pulling images for project: %s\n", project.Name)

	composeFilePath, err := utils.GetComposeFilePath(projectDir)
	if err != nil {
		return err
	}

	if err := cm.executeCommandWithErrorHandling(projectDir, "compose", "-f", composeFilePath, "pull"); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully pulled images for project: %s\n", project.Name)
	return nil
}

// BuildImages builds all images for the project
func (cm *ComposeManager) BuildImages(projectDir string, noBuildCache bool) error {
	// Check Docker health first
	if err := CheckDockerStatus(); err != nil {
		return err
	}

	project, err := cm.LoadProject(projectDir)
	if err != nil {
		return err
	}

	fmt.Printf("🔨 Building images for project: %s\n", project.Name)

	composeFilePath, err := utils.GetComposeFilePath(projectDir)
	if err != nil {
		return err
	}

	args := []string{"compose", "-f", composeFilePath, "build"}
	if noBuildCache {
		args = append(args, "--no-cache")
	}

	if err := cm.executeCommandWithErrorHandling(projectDir, args...); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully built images for project: %s\n", project.Name)
	return nil
}

// executeCommandWithErrorHandling executes docker commands with enhanced error handling
func (cm *ComposeManager) executeCommandWithErrorHandling(workingDir string, args ...string) error {
	cmd := exec.Command("docker", args...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		// Capture stderr for error analysis
		cmdForError := exec.Command("docker", args...)
		cmdForError.Dir = workingDir
		errorOutput, _ := cmdForError.CombinedOutput()
		errorStr := string(errorOutput)

		// Check for registry authentication errors
		if regError := DetectRegistryError(errorStr); regError != nil {
			return HandleRegistryError(regError, errorStr)
		}

		// Check if this is a Docker daemon connectivity issue
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			// If Docker daemon is not running, provide helpful error
			if strings.Contains(string(exitError.Stderr), "Cannot connect to the Docker daemon") ||
				strings.Contains(err.Error(), "connection refused") {
				fmt.Println()
				return fmt.Errorf("docker daemon is not running. Please start Docker Desktop and try again")
			}
		}

		// Check for other common Docker errors
		if strings.Contains(errorStr, "no such file or directory") {
			return fmt.Errorf("docker-compose file not found or invalid path")
		}

		if strings.Contains(errorStr, "network") && strings.Contains(errorStr, "already exists") {
			fmt.Println("⚠️  Network conflict detected - this usually resolves itself")
		}

		return err
	}

	return nil
}

// Helper function to execute docker commands (legacy - now uses enhanced version)
func (cm *ComposeManager) executeCommand(workingDir string, args ...string) error {
	return cm.executeCommandWithErrorHandling(workingDir, args...)
}

// Helper function to format port information
func (cm *ComposeManager) formatPorts(ports []dockertypes.Port) string {
	if len(ports) == 0 {
		return ""
	}

	var portStrings []string
	for _, port := range ports {
		if port.PublicPort > 0 {
			portStrings = append(portStrings, fmt.Sprintf("%d:%d", port.PublicPort, port.PrivatePort))
		} else {
			portStrings = append(portStrings, fmt.Sprintf("%d", port.PrivatePort))
		}
	}

	return strings.Join(portStrings, ", ")
}

// ExecuteDockerComposeCommand Legacy function for backward compatibility - now uses the Compose Manager
func ExecuteDockerComposeCommand(projectDir string, args ...string) error {
	cm, err := NewComposeManager()
	if err != nil {
		return fmt.Errorf("failed to create compose manager: %v", err)
	}
	defer cm.Close()

	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]
	restArgs := args[1:]

	switch command {
	case "up":
		detached := contains(restArgs, "-d") || contains(restArgs, "--detach")
		removeOrphans := contains(restArgs, "--remove-orphans")
		return cm.StartProject(projectDir, detached, removeOrphans)

	case "down":
		removeVolumes := contains(restArgs, "-v") || contains(restArgs, "--volumes")
		removeImages := contains(restArgs, "--rmi")
		return cm.StopProject(projectDir, removeVolumes, removeImages)

	case "restart":
		return cm.RestartProject(projectDir)

	case "pause":
		return cm.PauseProject(projectDir)

	case "unpause":
		return cm.UnpauseProject(projectDir)

	case "pull":
		return cm.PullImages(projectDir)

	case "build":
		noBuildCache := contains(restArgs, "--no-cache")
		return cm.BuildImages(projectDir, noBuildCache)

	case "logs":
		follow := contains(restArgs, "-f") || contains(restArgs, "--follow")
		var services []string
		// Extract service names from args
		for _, arg := range restArgs {
			if !strings.HasPrefix(arg, "-") {
				services = append(services, arg)
			}
		}
		return cm.ViewLogs(projectDir, services, follow)

	default:
		// For unsupported commands, fall back to direct execution
		return cm.executeCommandWithErrorHandling(projectDir, append([]string{"compose", "-f"}, args...)...)
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

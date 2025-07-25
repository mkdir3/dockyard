package docker

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/docker/client"
)

// HealthChecker handles Docker daemon connectivity and health checks
type HealthChecker struct {
	client client.APIClient
	ctx    context.Context
}

// NewDockerHealthChecker creates a new Docker health checker
func NewDockerHealthChecker() (*HealthChecker, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &HealthChecker{
		client: cli,
		ctx:    ctx,
	}, nil
}

// Close closes the Docker client connection
func (dhc *HealthChecker) Close() error {
	if dhc.client != nil {
		return dhc.client.Close()
	}
	return nil
}

// CheckDockerDaemon checks if Docker daemon is running and accessible
func (dhc *HealthChecker) CheckDockerDaemon() error {
	// Set a timeout for the ping
	ctx, cancel := context.WithTimeout(dhc.ctx, 5*time.Second)
	defer cancel()

	// Try to ping Docker daemon
	_, err := dhc.client.Ping(ctx)
	return err
}

// IsDockerDesktopInstalled checks if Docker Desktop is installed (macOS/Windows)
func IsDockerDesktopInstalled() bool {
	switch runtime.GOOS {
	case "darwin": // macOS
		_, err := exec.LookPath("docker")
		if err != nil {
			return false
		}
		// Check if Docker Desktop app exists
		_, err = exec.Command("ls", "/Applications/Docker.app").Output()
		return err == nil
	case "windows":
		_, err := exec.LookPath("docker")
		return err == nil
	default: // Linux
		_, err := exec.LookPath("docker")
		return err == nil
	}
}

// StartDockerDesktop attempts to start Docker Desktop (macOS only for now)
func StartDockerDesktop() error {
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd := exec.Command("open", "-a", "Docker")
		return cmd.Run()
	default:
		return fmt.Errorf("automatic Docker startup not supported on %s", runtime.GOOS)
	}
}

// CheckDockerStatus performs comprehensive Docker status check
func CheckDockerStatus() error {
	fmt.Println("🔍 Checking Docker status...")

	// Check if Docker is installed
	if !IsDockerDesktopInstalled() {
		return fmt.Errorf("docker is not installed. Please install Docker Desktop from https://www.docker.com/products/docker-desktop")
	}

	// Create health checker
	dhc, err := NewDockerHealthChecker()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer func(dhc *HealthChecker) {
		err := dhc.Close()
		if err != nil {
			fmt.Printf("❌ Failed to close Docker client: %v\n", err)
		} else {
			fmt.Println("✅ Docker client connection closed")
		}
	}(dhc)

	// Check if daemon is running
	err = dhc.CheckDockerDaemon()
	if err != nil {
		return handleDockerDaemonError(err)
	}

	fmt.Println("✅ Docker daemon is running and accessible")
	return nil
}

// handleDockerDaemonError provides user-friendly error handling and recovery options
func handleDockerDaemonError(err error) error {
	fmt.Printf("❌ Docker daemon is not accessible: %v\n\n", err)

	// Provide platform-specific guidance
	switch runtime.GOOS {
	case "darwin": // macOS
		return handleMacOSDockerError()
	case "windows":
		return handleWindowsDockerError()
	default: // Linux
		return handleLinuxDockerError()
	}
}

// handleMacOSDockerError handles Docker issues on macOS
func handleMacOSDockerError() error {
	fmt.Println("🍎 macOS Docker troubleshooting:")
	fmt.Println("   1. Docker Desktop might not be running")
	fmt.Println("   2. Check if Docker Desktop is installed in /Applications/")
	fmt.Println("   3. Docker Desktop might be starting up (this can take a few minutes)")
	fmt.Println()

	var action string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"Try to start Docker Desktop automatically",
			"Wait and retry (Docker might be starting)",
			"Get manual instructions",
			"Exit and fix manually",
		},
	}

	err := survey.AskOne(prompt, &action)
	if err != nil {
		return err
	}

	switch action {
	case "Try to start Docker Desktop automatically":
		return attemptDockerDesktopStart()
	case "Wait and retry (Docker might be starting)":
		return waitAndRetryDocker()
	case "Get manual instructions":
		return showManualInstructions()
	default:
		return fmt.Errorf("please start Docker Desktop manually and try again")
	}
}

// handleWindowsDockerError handles Docker issues on Windows
func handleWindowsDockerError() error {
	fmt.Println("🪟 Windows Docker troubleshooting:")
	fmt.Println("   1. Start Docker Desktop from the Start menu")
	fmt.Println("   2. Wait for Docker Desktop to fully start (check system tray)")
	fmt.Println("   3. Ensure WSL 2 is properly configured (if using WSL 2 backend)")
	fmt.Println()
	return fmt.Errorf("please start Docker Desktop manually and try again")
}

// handleLinuxDockerError handles Docker issues on Linux
func handleLinuxDockerError() error {
	fmt.Println("🐧 Linux Docker troubleshooting:")
	fmt.Println("   1. Start Docker daemon: sudo systemctl start docker")
	fmt.Println("   2. Enable Docker service: sudo systemctl enable docker")
	fmt.Println("   3. Add user to docker group: sudo usermod -aG docker $USER")
	fmt.Println("   4. Log out and back in for group changes to take effect")
	fmt.Println()
	return fmt.Errorf("please start Docker daemon manually and try again")
}

// attemptDockerDesktopStart tries to start Docker Desktop automatically
func attemptDockerDesktopStart() error {
	fmt.Println("🚀 Attempting to start Docker Desktop...")

	err := StartDockerDesktop()
	if err != nil {
		fmt.Printf("❌ Failed to start Docker Desktop automatically: %v\n", err)
		return showManualInstructions()
	}

	fmt.Println("✅ Docker Desktop start command sent!")
	return waitAndRetryDocker()
}

// waitAndRetryDocker waits for Docker to start and retries the connection
func waitAndRetryDocker() error {
	fmt.Println("⏳ Waiting for Docker to start...")
	fmt.Println("   This can take many seconds for Docker Desktop to fully initialize...")

	maxRetries := 12 // 60 seconds total
	for i := 0; i < maxRetries; i++ {
		time.Sleep(5 * time.Second)

		dhc, err := NewDockerHealthChecker()
		if err != nil {
			continue
		}

		err = dhc.CheckDockerDaemon()
		dhc.Close()

		if err == nil {
			fmt.Println("✅ Docker is now running!")
			return nil
		}

		dots := strings.Repeat(".", (i%3)+1)
		fmt.Printf("   Still waiting%s (%d/%d)\r", dots, i+1, maxRetries)
	}

	fmt.Println()
	fmt.Println("❌ Docker failed to start within 60 seconds")
	return showManualInstructions()
}

// showManualInstructions shows manual troubleshooting steps
func showManualInstructions() error {
	fmt.Println()
	fmt.Println("📖 Manual troubleshooting steps:")
	fmt.Println()

	switch runtime.GOOS {
	case "darwin":
		fmt.Println("   1. Open Docker Desktop from Applications folder")
		fmt.Println("   2. Wait for the Docker whale icon to appear in the menu bar")
		fmt.Println("   3. Click the whale icon and ensure it shows 'Docker Desktop is running'")
		fmt.Println("   4. If Docker Desktop won't start, try restarting your Mac")
	case "windows":
		fmt.Println("   1. Search for 'Docker Desktop' in the Start menu and launch it")
		fmt.Println("   2. Wait for the Docker whale icon in the system tray")
		fmt.Println("   3. Ensure WSL 2 is enabled if using WSL 2 backend")
		fmt.Println("   4. Try running 'docker --version' in Command Prompt")
	default:
		fmt.Println("   1. sudo systemctl start docker")
		fmt.Println("   2. sudo systemctl enable docker")
		fmt.Println("   3. sudo usermod -aG docker $USER")
		fmt.Println("   4. Log out and back in")
	}

	fmt.Println()
	fmt.Println("   Then run dockyard again!")
	fmt.Println()

	return fmt.Errorf("please follow the manual steps above and try again")
}

//TODO: stop docker daemon

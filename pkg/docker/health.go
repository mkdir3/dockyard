package docker

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"dockyard/pkg/ui"
	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/docker/client"
)

const (
	PingTimeout         = 5 * time.Second
	RuntimeStartTimeout = 60 * time.Second
	RetryInterval       = 5 * time.Second
	MaxRetries          = 12
)

type ContainerRuntime string

type Platform string

const (
	PlatformDarwin  Platform = "darwin"
	PlatformWindows Platform = "windows"
	PlatformLinux   Platform = "linux"
)

const (
	CommandOrbctl = "orbctl"
	CommandColima = "colima"
	CommandDocker = "docker"
	CommandPodman = "podman"
)

type ContainerRuntimeError struct {
	Runtime ContainerRuntime
	Err     error
}

func (e *ContainerRuntimeError) Error() string {
	return fmt.Sprintf("container runtime %s error: %v", e.Runtime, e.Err)
}

func (e *ContainerRuntimeError) Unwrap() error {
	return e.Err
}

type RuntimeNotFoundError struct {
	Platform Platform
}

func (e *RuntimeNotFoundError) Error() string {
	return fmt.Sprintf("no container runtime found for platform %s", e.Platform)
}

type HealthChecker struct {
	client client.APIClient
	ctx    context.Context
}

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

func (dhc *HealthChecker) Close() error {
	if dhc.client != nil {
		return dhc.client.Close()
	}
	return nil
}

func (dhc *HealthChecker) CheckDockerDaemon() error {
	ctx, cancel := context.WithTimeout(dhc.ctx, PingTimeout)
	defer cancel()

	_, err := dhc.client.Ping(ctx)
	return err
}

func IsDockerAvailable() bool {
	_, err := exec.LookPath(CommandDocker)
	return err == nil
}

func StartDockerDesktop() error {
	switch runtime.GOOS {
	case string(PlatformDarwin):
		cmd := exec.Command("open", "-a", "Docker")
		return cmd.Run()
	default:
		return fmt.Errorf("automatic Docker startup not supported on %s", runtime.GOOS)
	}
}

func CheckDockerStatus() error {
	fmt.Println(ui.RenderInfo(config.Common.DockerStatusCheck))

	if !IsDockerAvailable() {
		return handleDockerNotInstalled()
	}

	dhc, err := NewDockerHealthChecker()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer func(dhc *HealthChecker) {
		err := dhc.Close()
		if err != nil {
			fmt.Printf("❌ Failed to close Docker client: %v\n", err)
		}
	}(dhc)

	err = dhc.CheckDockerDaemon()
	if err != nil {
		return handleDockerDaemonError(err)
	}

	fmt.Println(ui.RenderSuccess(config.Common.DockerRunning))
	return nil
}

func handleDockerNotInstalled() error {
	fmt.Println(ui.RenderError(config.Common.DockerNotFound))

	platformConfig := getPlatformConfiguration(runtime.GOOS)
	printLines(platformConfig.InstallOptions)

	fmt.Println()
	return fmt.Errorf(config.ErrorMessages.InstallRuntime)
}

func handleDockerDaemonError(err error) error {
	fmt.Printf("❌ Docker daemon is not accessible: %v\n\n", err)

	switch runtime.GOOS {
	case string(PlatformDarwin):
		return handleMacOSDockerError()
	case string(PlatformWindows):
		return handleWindowsDockerError()
	default: // Linux
		return handleLinuxDockerError()
	}
}

func handleMacOSDockerError() error {
	platformConfig := getPlatformConfiguration(runtime.GOOS)
	printLines(platformConfig.Troubleshooting)
	fmt.Println()

	var action string
	prompt := &survey.Select{
		Message: config.UIOptions.RuntimeOptionsMessage,
		Options: config.UIOptions.RuntimeOptions,
	}

	err := survey.AskOne(prompt, &action)
	if err != nil {
		return err
	}

	switch action {
	case config.UIOptions.RuntimeOptions[0]: // "Try to start container runtime automatically"
		return attemptContainerRuntimeStart()
	case config.UIOptions.RuntimeOptions[1]: // "Wait and retry (container runtime might be starting)"
		return waitAndRetryDocker()
	case config.UIOptions.RuntimeOptions[2]: // "Get manual startup instructions"
		return showStartupOptions()
	default:
		return fmt.Errorf(config.ErrorMessages.StartRuntimeManually)
	}
}

func handleWindowsDockerError() error {
	platformConfig := getPlatformConfiguration("windows")
	printLines(platformConfig.Troubleshooting)
	fmt.Println()
	return fmt.Errorf(config.ErrorMessages.DockerDesktopManual)
}

func handleLinuxDockerError() error {
	platformConfig := getPlatformConfiguration("linux")
	printLines(platformConfig.Troubleshooting)
	fmt.Println()
	return fmt.Errorf(config.ErrorMessages.DockerDaemonManual)
}

func attemptOrbStackStart() error {
	cmd := exec.Command("open", "-a", "OrbStack")
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to start OrbStack automatically: %v\n", err)
		return showOrbStackInstructions()
	}

	fmt.Println(ui.RenderSuccess(config.Common.OrbStackStartSent))
	fmt.Println(ui.RenderInfo(config.Common.OrbStackNote))
	return waitAndRetryDocker()
}

func showOrbStackInstructions() error {
	platformConfig := getPlatformConfiguration(runtime.GOOS)
	if orbInstructions, exists := platformConfig.Runtimes["orbstack"]; exists {
		printLines(orbInstructions.ManualStart)
		fmt.Println()
		printLines(orbInstructions.AutoStart)
		fmt.Println()
	}
	return fmt.Errorf(config.ErrorMessages.StartOrbStack)
}

func attemptColimaStart() error {
	cmd := exec.Command(CommandColima, "start")
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to start Colima: %v\n", err)
		return showColimaInstructions()
	}

	fmt.Println(ui.RenderSuccess(config.Common.ColimaStartSent))
	fmt.Println(ui.RenderInfo(config.Common.ColimaNote))
	return waitAndRetryDocker()
}

func showColimaInstructions() error {
	platformConfig := getPlatformConfiguration(runtime.GOOS)
	if colimaInstructions, exists := platformConfig.Runtimes["colima"]; exists {
		printLines(colimaInstructions.ManualStart)
		fmt.Println()
		printLines(colimaInstructions.AutoStart)
		fmt.Println()
		printLines(colimaInstructions.Commands)
		fmt.Println()
	}
	return fmt.Errorf(config.ErrorMessages.StartColima)
}

func attemptContainerRuntimeStart() error {
	fmt.Println(ui.RenderInfo(config.Common.RuntimeStartAttempt))

	if runtime.GOOS == string(PlatformDarwin) {
		if _, err := exec.LookPath(CommandOrbctl); err == nil {
			fmt.Println(ui.RenderInfo("   Found " + ui.RenderRuntimeIcon("orbstack") + " OrbStack, attempting to start..."))
			return attemptOrbStackStart()
		}

		if _, err := exec.LookPath(CommandColima); err == nil {
			fmt.Println(ui.RenderInfo("   Found " + ui.RenderRuntimeIcon("colima") + " Colima, attempting to start..."))
			return attemptColimaStart()
		}

		err := StartDockerDesktop()
		if err != nil {
			fmt.Printf("❌ Failed to start container runtime automatically: %v\n", err)
			return showStartupOptions()
		}

		fmt.Println(ui.RenderSuccess(config.Common.DockerDesktopSent))
		return waitAndRetryDocker()
	}

	err := StartDockerDesktop()
	if err != nil {
		fmt.Printf("❌ Failed to start container runtime automatically: %v\n", err)
		return showStartupOptions()
	}

	fmt.Println(ui.RenderSuccess(config.Common.ContainerRuntimeSent))
	return waitAndRetryDocker()
}

func waitAndRetryDocker() error {
	fmt.Println(ui.RenderInfo(config.Common.RuntimeWaiting))

	for i := 0; i < MaxRetries; i++ {
		time.Sleep(RetryInterval)

		dhc, err := NewDockerHealthChecker()
		if err != nil {
			continue
		}

		err = dhc.CheckDockerDaemon()
		dhc.Close()

		if err == nil {
			fmt.Println(ui.RenderSuccess("Container runtime is now running!"))
			return nil
		}

		dots := strings.Repeat(".", (i%3)+1)
		fmt.Printf("   Still waiting%s (%d/%d)\r", dots, i+1, MaxRetries)
	}

	fmt.Println()
	fmt.Printf(ui.RenderError(fmt.Sprintf(config.Common.RuntimeStartFailed, int(RuntimeStartTimeout.Seconds()))))
	return showStartupOptions()
}

func showStartupOptions() error {
	fmt.Println()

	var choice string
	prompt := &survey.Select{
		Message: config.UIOptions.StartupOptionsMessage,
		Options: config.UIOptions.StartupOptions,
	}

	err := survey.AskOne(prompt, &choice)
	if err != nil {
		return err
	}

	switch choice {
	case config.UIOptions.StartupOptions[0]: // "Show manual startup commands"
		return showManualStartup()
	case config.UIOptions.StartupOptions[1]: // "Show auto-start setup (start with computer)"
		return showAutoStartSetup()
	default:
		return fmt.Errorf(config.ErrorMessages.StartRuntimeManually)
	}
}

func showManualStartup() error {
	msgs := getStartupInstructions(runtime.GOOS, "manual")
	printLines(msgs)
	fmt.Println()
	return fmt.Errorf(config.ErrorMessages.ManualStartup)
}

func showAutoStartSetup() error {
	msgs := getStartupInstructions(runtime.GOOS, "auto")
	printLines(msgs)
	fmt.Println()
	return fmt.Errorf(config.ErrorMessages.AutoStartSetup)
}

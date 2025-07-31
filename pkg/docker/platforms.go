package docker

import (
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
)

//go:embed platforms.yaml
var platformsYAML []byte

type RuntimeInstructions struct {
	ManualStart []string `yaml:"manual_start"`
	AutoStart   []string `yaml:"auto_start"`
	Commands    []string `yaml:"commands"`
}

type StartupInstructions struct {
	Manual []string `yaml:"manual"`
	Auto   []string `yaml:"auto"`
}

type PlatformConfiguration struct {
	InstallOptions  []string                       `yaml:"install_options"`
	Troubleshooting []string                       `yaml:"troubleshooting"`
	Runtimes        map[string]RuntimeInstructions `yaml:"runtimes"`
	Startup         StartupInstructions            `yaml:"startup"`
}

type CommonMessages struct {
	DockerNotFound       string `yaml:"docker_not_found"`
	DockerStatusCheck    string `yaml:"docker_status_check"`
	DockerRunning        string `yaml:"docker_running"`
	RuntimeStartAttempt  string `yaml:"runtime_start_attempt"`
	RuntimeWaiting       string `yaml:"runtime_waiting"`
	RuntimeStartFailed   string `yaml:"runtime_start_failed"`
	OrbStackStartSent    string `yaml:"orbstack_start_sent"`
	OrbStackNote         string `yaml:"orbstack_note"`
	ColimaStartSent      string `yaml:"colima_start_sent"`
	ColimaNote           string `yaml:"colima_note"`
	DockerDesktopSent    string `yaml:"docker_desktop_sent"`
	ContainerRuntimeSent string `yaml:"container_runtime_sent"`
}

type UIOptions struct {
	StartupOptionsMessage string   `yaml:"startup_options_message"`
	StartupOptions        []string `yaml:"startup_options"`
	RuntimeOptionsMessage string   `yaml:"runtime_options_message"`
	RuntimeOptions        []string `yaml:"runtime_options"`
}

type ErrorMessages struct {
	InstallRuntime       string `yaml:"install_runtime"`
	StartRuntimeManually string `yaml:"start_runtime_manually"`
	StartOrbStack        string `yaml:"start_orbstack"`
	StartColima          string `yaml:"start_colima"`
	ManualStartup        string `yaml:"manual_startup"`
	AutoStartSetup       string `yaml:"auto_start_setup"`
	DockerDesktopManual  string `yaml:"docker_desktop_manual"`
	DockerDaemonManual   string `yaml:"docker_daemon_manual"`
}

type PlatformsConfig struct {
	Common        CommonMessages                   `yaml:"common"`
	Platforms     map[string]PlatformConfiguration `yaml:"platforms"`
	UIOptions     UIOptions                        `yaml:"ui_options"`
	ErrorMessages ErrorMessages                    `yaml:"error_messages"`
}

var config PlatformsConfig

func init() {
	if err := yaml.Unmarshal(platformsYAML, &config); err != nil {
		panic(fmt.Sprintf("Failed to load platforms config from platforms.yaml: %v\nPlease check that platforms.yaml exists, is accessible, and is valid YAML format.", err))
	}
}

func printLines(lines []string) {
	for _, line := range lines {
		if line == "" {
			fmt.Println()
		} else {
			fmt.Println(line)
		}
	}
}

func getPlatformConfiguration(platform string) PlatformConfiguration {
	if platformConfig, exists := config.Platforms[platform]; exists {
		return platformConfig
	}
	return config.Platforms["linux"] // fallback
}

func getStartupInstructions(platform, instructionType string) []string {
	platformConfig := getPlatformConfiguration(platform)
	switch instructionType {
	case "manual":
		return platformConfig.Startup.Manual
	case "auto":
		return platformConfig.Startup.Auto
	default:
		return platformConfig.Startup.Manual
	}
}

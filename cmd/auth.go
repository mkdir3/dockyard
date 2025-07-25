package cmd

import (
	"dockyard/pkg/docker"
	"fmt"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Docker registries",
	Long:  `Interactive helper to authenticate with various Docker registries like GitLab, GitHub, Docker Hub, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		runAuthWizard()
	},
}

func runAuthWizard() {
	fmt.Println("🔐 Docker Registry Authentication Wizard")
	fmt.Println("========================================")
	fmt.Println()

	// Check if Docker is running
	if err := docker.CheckDockerStatus(); err != nil {
		fmt.Printf("❌ Docker check failed: %v\n", err)
		return
	}

	// Registry selection
	var registry string
	registryPrompt := &survey.Select{
		Message: "Which registry do you want to authenticate with?",
		Options: []string{
			"GitLab Container Registry (registry.gitlab.com)",
			"GitHub Container Registry (ghcr.io)",
			"Docker Hub (docker.io)",
			"Custom Registry",
			"Check current authentication status",
		},
	}

	err := survey.AskOne(registryPrompt, &registry)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	switch registry {
	case "GitLab Container Registry (registry.gitlab.com)":
		authenticateGitLab()
	case "GitHub Container Registry (ghcr.io)":
		authenticateGitHub()
	case "Docker Hub (docker.io)":
		authenticateDockerHub()
	case "Custom Registry":
		authenticateCustomRegistry()
	case "Check current authentication status":
		checkAuthStatus()
	}
}

func authenticateGitLab() {
	fmt.Println("\n🦊 GitLab Container Registry Authentication")
	fmt.Println("==========================================")

	showGitLabInstructions()

	var proceed string
	proceedPrompt := &survey.Select{
		Message: "Have you created a GitLab Personal Access Token?",
		Options: []string{"Yes, I have a token", "No, help me create one", "Cancel"},
	}

	err := survey.AskOne(proceedPrompt, &proceed)
	if err != nil {
		return
	}

	switch proceed {
	case "Yes, I have a token":
		performLogin("registry.gitlab.com")
	case "No, help me create one":
		fmt.Println("\n📖 Step-by-step token creation:")
		fmt.Println("1. Open: https://gitlab.com/-/profile/personal_access_tokens")
		fmt.Println("2. Click 'Add new token'")
		fmt.Println("3. Set name: 'Docker Registry Access'")
		fmt.Println("4. Select scope: 'read_registry' ✅")
		fmt.Println("5. Click 'Create personal access token'")
		fmt.Println("6. Copy the token (you won't see it again!)")
		fmt.Println()

		var ready string
		survey.AskOne(&survey.Select{
			Message: "Ready to login?",
			Options: []string{"Yes, login now", "I'll do it later"},
		}, &ready)

		if ready == "Yes, login now" {
			performLogin("registry.gitlab.com")
		}
	}
}

func authenticateGitHub() {
	fmt.Println("\n🐙 GitHub Container Registry Authentication")
	fmt.Println("==========================================")

	showGitHubInstructions()

	var proceed string
	proceedPrompt := &survey.Select{
		Message: "Have you created a GitHub Personal Access Token?",
		Options: []string{"Yes, I have a token", "No, help me create one", "Cancel"},
	}

	err := survey.AskOne(proceedPrompt, &proceed)
	if err != nil {
		return
	}

	switch proceed {
	case "Yes, I have a token":
		performLogin("ghcr.io")
	case "No, help me create one":
		fmt.Println("\n📖 Step-by-step token creation:")
		fmt.Println("1. Open: https://github.com/settings/tokens")
		fmt.Println("2. Click 'Generate new token (classic)'")
		fmt.Println("3. Set name: 'Docker Registry Access'")
		fmt.Println("4. Select scope: 'read:packages' ✅")
		fmt.Println("5. Click 'Generate token'")
		fmt.Println("6. Copy the token immediately!")
		fmt.Println()

		var ready string
		survey.AskOne(&survey.Select{
			Message: "Ready to login?",
			Options: []string{"Yes, login now", "I'll do it later"},
		}, &ready)

		if ready == "Yes, login now" {
			performLogin("ghcr.io")
		}
	}
}

func authenticateDockerHub() {
	fmt.Println("\n🐳 Docker Hub Authentication")
	fmt.Println("============================")

	fmt.Println("Docker Hub supports both password and access token authentication.")
	fmt.Println("Access tokens are recommended for better security.")
	fmt.Println()

	var authMethod string
	survey.AskOne(&survey.Select{
		Message: "What authentication method do you prefer?",
		Options: []string{"Username & Password", "Username & Access Token", "Cancel"},
	}, &authMethod)

	switch authMethod {
	case "Username & Password":
		performLogin("")
	case "Username & Access Token":
		fmt.Println("\n📖 Creating a Docker Hub Access Token:")
		fmt.Println("1. Go to: https://hub.docker.com/settings/security")
		fmt.Println("2. Click 'New Access Token'")
		fmt.Println("3. Enter description: 'Docker Manager CLI'")
		fmt.Println("4. Set permissions as needed")
		fmt.Println("5. Click 'Generate'")
		fmt.Println("6. Copy the token")
		fmt.Println()
		performLogin("")
	}
}

func authenticateCustomRegistry() {
	fmt.Println("\n🌐 Custom Registry Authentication")
	fmt.Println("================================")

	var registryURL string
	err := survey.AskOne(&survey.Input{
		Message: "Enter the registry URL (e.g., my-registry.com):",
	}, &registryURL)
	if err != nil {
		return
	}

	performLogin(registryURL)
}

func performLogin(registryURL string) {
	fmt.Printf("\n🔑 Logging in to %s\n", getRegistryDisplayName(registryURL))

	var username string
	err := survey.AskOne(&survey.Input{
		Message: "Username:",
	}, &username)
	if err != nil {
		return
	}

	var password string
	err = survey.AskOne(&survey.Password{
		Message: "Password/Token:",
	}, &password)
	if err != nil {
		return
	}

	// Perform docker login
	fmt.Println("\n🔐 Authenticating...")

	var cmd *exec.Cmd
	if registryURL == "" {
		// Docker Hub
		cmd = exec.Command("docker", "login", "-u", username, "--password-stdin")
	} else {
		cmd = exec.Command("docker", "login", registryURL, "-u", username, "--password-stdin")
	}

	cmd.Stdin = strings.NewReader(password)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("❌ Login failed: %s\n", string(output))
		return
	}

	fmt.Printf("✅ Successfully authenticated with %s!\n", getRegistryDisplayName(registryURL))
	fmt.Println("🎉 You can now pull private images from this registry.")
}

func checkAuthStatus() {
	fmt.Println("\n🔍 Checking Docker authentication status...")
	fmt.Println("==========================================")

	// Check if user is logged in to Docker Hub
	cmd := exec.Command("docker", "info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to get Docker info: %v\n", err)
		return
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Username:") {
		fmt.Println("✅ Authenticated with Docker Hub")
	} else {
		fmt.Println("❌ Not authenticated with Docker Hub")
	}

	// Try to get registry auth info from Docker config
	fmt.Println("\n📋 Checking configured registries...")

	// This is a simple check - in practice, you might want to read ~/.docker/config.json
	registries := []string{"registry.gitlab.com", "ghcr.io"}

	for _, registry := range registries {
		cmd := exec.Command("docker", "login", registry, "--get-login")
		_, err := cmd.CombinedOutput()
		if err == nil {
			fmt.Printf("✅ Configured: %s\n", registry)
		} else {
			fmt.Printf("❌ Not configured: %s\n", registry)
		}
	}

	fmt.Println("\n💡 Tip: Use 'dockyard auth' to set up authentication for private registries.")
}

func getRegistryDisplayName(registryURL string) string {
	switch registryURL {
	case "registry.gitlab.com":
		return "GitLab Container Registry"
	case "ghcr.io":
		return "GitHub Container Registry"
	case "":
		return "Docker Hub"
	default:
		return registryURL
	}
}

func showGitLabInstructions() {
	fmt.Println("GitLab requires a Personal Access Token for registry access.")
	fmt.Println("📋 Required scope: read_registry")
	fmt.Println("🌐 Token URL: https://gitlab.com/-/profile/personal_access_tokens")
	fmt.Println()
}

func showGitHubInstructions() {
	fmt.Println("GitHub requires a Personal Access Token for container registry access.")
	fmt.Println("📋 Required scope: read:packages")
	fmt.Println("🌐 Token URL: https://github.com/settings/tokens")
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(authCmd)
}

package docker

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"os/exec"
	"regexp"
	"strings"
)

// RegistryError represents different types of registry authentication errors
type RegistryError struct {
	Registry    string
	ErrorType   string
	Image       string
	Suggestions []string
}

// DetectRegistryError analyzes error output to identify registry authentication issues
func DetectRegistryError(errorOutput string) *RegistryError {
	// Common registry error patterns
	patterns := map[string]*regexp.Regexp{
		"gitlab_auth":     regexp.MustCompile(`error from registry.*gitlab\.com.*HTTP Basic.*Access denied`),
		"github_auth":     regexp.MustCompile(`error from registry.*ghcr\.io.*unauthorized`),
		"dockerhub_auth":  regexp.MustCompile(`pull access denied.*repository does not exist or may require.*docker login`),
		"generic_auth":    regexp.MustCompile(`unauthorized.*authentication required`),
		"registry_access": regexp.MustCompile(`error from registry.*Access denied`),
	}

	// Extract registry and image info
	registryMatch := regexp.MustCompile(`registry\.([a-zA-Z0-9.-]+)`).FindStringSubmatch(errorOutput)
	imageMatch := regexp.MustCompile(`unable to get image '([^']+)'`).FindStringSubmatch(errorOutput)

	var registry, image string
	if len(registryMatch) > 1 {
		registry = registryMatch[1]
	}
	if len(imageMatch) > 1 {
		image = imageMatch[1]
	}

	// Determine error type and provide specific suggestions
	for errorType, pattern := range patterns {
		if pattern.MatchString(errorOutput) {
			return &RegistryError{
				Registry:    registry,
				ErrorType:   errorType,
				Image:       image,
				Suggestions: getRegistrySuggestions(errorType, registry),
			}
		}
	}

	return nil
}

// getRegistrySuggestions provides specific suggestions based on registry type
func getRegistrySuggestions(errorType, registry string) []string {
	switch errorType {
	case "gitlab_auth":
		return []string{
			"Create a GitLab Personal Access Token with 'read_registry' scope",
			fmt.Sprintf("Run: docker login %s", getRegistryURL(registry)),
			"Use your GitLab username and the Personal Access Token as password",
			"GitLab tokens: https://gitlab.com/-/profile/personal_access_tokens",
		}
	case "github_auth":
		return []string{
			"Create a GitHub Personal Access Token with 'read:packages' scope",
			"Run: docker login ghcr.io",
			"Use your GitHub username and the Personal Access Token as password",
			"GitHub tokens: https://github.com/settings/tokens",
		}
	case "dockerhub_auth":
		return []string{
			"Run: docker login",
			"Use your Docker Hub username and password",
			"Or create an Access Token in Docker Hub settings",
		}
	default:
		return []string{
			"Run: docker login <registry-url>",
			"Use appropriate credentials for the registry",
			"Check if the image exists and you have permission to access it",
		}
	}
}

// getRegistryURL returns the full registry URL for login
func getRegistryURL(registry string) string {
	if strings.Contains(registry, "gitlab.com") {
		return "registry.gitlab.com"
	}
	if strings.Contains(registry, "github.com") || strings.Contains(registry, "ghcr.io") {
		return "ghcr.io"
	}
	return registry
}

// HandleRegistryError provides interactive assistance for registry authentication
func HandleRegistryError(regError *RegistryError, errorOutput string) error {
	fmt.Println()
	fmt.Printf("🔐 Docker Registry Authentication Error Detected!\n")
	fmt.Printf("📦 Image: %s\n", regError.Image)
	fmt.Printf("🌐 Registry: %s\n", regError.Registry)
	fmt.Println()

	// Show specific error details
	fmt.Println("📋 Error Details:")
	if strings.Contains(errorOutput, "HTTP Basic: Access denied") {
		fmt.Println("   • Authentication failed - invalid credentials")
	}
	if strings.Contains(errorOutput, "token was either incorrect, expired, or improperly scoped") {
		fmt.Println("   • Token issue - check token validity and permissions")
	}
	if strings.Contains(errorOutput, "password was incorrect") {
		fmt.Println("   • Password authentication failed")
	}
	fmt.Println()

	// Show suggestions
	fmt.Println("💡 How to fix this:")
	for i, suggestion := range regError.Suggestions {
		fmt.Printf("   %d. %s\n", i+1, suggestion)
	}
	fmt.Println()

	// Offer interactive assistance
	var action string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"Help me login to the registry",
			"Show detailed authentication guide",
			"Skip this project for now",
			"Open registry documentation",
		},
	}

	err := survey.AskOne(prompt, &action)
	if err != nil {
		return err
	}

	switch action {
	case "Help me login to the registry":
		return assistWithLogin(regError)
	case "Show detailed authentication guide":
		return showDetailedGuide(regError)
	case "Skip this project for now":
		fmt.Println("⏭️  Skipping this project. You can try again after authentication.")
		return fmt.Errorf("registry authentication required - skipped")
	case "Open registry documentation":
		return openRegistryDocs(regError)
	default:
		return fmt.Errorf("registry authentication required")
	}
}

// assistWithLogin helps the user login to the registry
func assistWithLogin(regError *RegistryError) error {
	registryURL := getRegistryURL(regError.Registry)

	fmt.Printf("\n🔑 Let's authenticate with %s\n\n", registryURL)

	// Get username
	var username string
	err := survey.AskOne(&survey.Input{
		Message: "Enter your username:",
	}, &username)
	if err != nil {
		return err
	}

	// Get password/token
	var password string
	err = survey.AskOne(&survey.Password{
		Message: "Enter your password or access token:",
	}, &password)
	if err != nil {
		return err
	}

	// Attempt login
	fmt.Printf("🔐 Attempting to login to %s...\n", registryURL)

	cmd := exec.Command("docker", "login", registryURL, "-u", username, "--password-stdin")
	cmd.Stdin = strings.NewReader(password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Login failed: %s\n", string(output))
		return fmt.Errorf("docker login failed: %v", err)
	}

	fmt.Printf("✅ Successfully logged in to %s!\n", registryURL)
	fmt.Println("💡 You can now retry starting your project.")
	return nil
}

// showDetailedGuide shows comprehensive authentication instructions
func showDetailedGuide(regError *RegistryError) error {
	fmt.Println()
	fmt.Printf("📖 Detailed Authentication Guide for %s\n", regError.Registry)
	fmt.Println(strings.Repeat("=", 50))

	switch regError.ErrorType {
	case "gitlab_auth":
		showGitLabGuide()
	case "github_auth":
		showGitHubGuide()
	case "dockerhub_auth":
		showDockerHubGuide()
	default:
		showGenericGuide(regError.Registry)
	}

	return fmt.Errorf("please follow the authentication steps above")
}

// showGitLabGuide shows GitLab-specific authentication guide
func showGitLabGuide() {
	fmt.Println()
	fmt.Println("🦊 GitLab Container Registry Authentication:")
	fmt.Println()
	fmt.Println("1. Create a Personal Access Token:")
	fmt.Println("   • Go to: https://gitlab.com/-/profile/personal_access_tokens")
	fmt.Println("   • Click 'Add new token'")
	fmt.Println("   • Name: 'Docker Registry Access'")
	fmt.Println("   • Scopes: ✅ read_registry (required)")
	fmt.Println("   • Expiration: Set as needed")
	fmt.Println("   • Click 'Create personal access token'")
	fmt.Println("   • 💾 SAVE THE TOKEN - you won't see it again!")
	fmt.Println()
	fmt.Println("2. Login to GitLab Registry:")
	fmt.Println("   docker login registry.gitlab.com")
	fmt.Println("   Username: <your-gitlab-username>")
	fmt.Println("   Password: <your-personal-access-token>")
	fmt.Println()
	fmt.Println("3. Verify access:")
	fmt.Println("   docker pull <your-image-name>")
}

// showGitHubGuide shows GitHub-specific authentication guide
func showGitHubGuide() {
	fmt.Println()
	fmt.Println("🐙 GitHub Container Registry Authentication:")
	fmt.Println()
	fmt.Println("1. Create a Personal Access Token:")
	fmt.Println("   • Go to: https://github.com/settings/tokens")
	fmt.Println("   • Click 'Generate new token (classic)'")
	fmt.Println("   • Name: 'Docker Registry Access'")
	fmt.Println("   • Scopes: ✅ read:packages (required)")
	fmt.Println("   • Click 'Generate token'")
	fmt.Println("   • 💾 COPY THE TOKEN immediately!")
	fmt.Println()
	fmt.Println("2. Login to GitHub Registry:")
	fmt.Println("   docker login ghcr.io")
	fmt.Println("   Username: <your-github-username>")
	fmt.Println("   Password: <your-personal-access-token>")
}

// showDockerHubGuide shows Docker Hub authentication guide
func showDockerHubGuide() {
	fmt.Println()
	fmt.Println("🐳 Docker Hub Authentication:")
	fmt.Println()
	fmt.Println("1. Login to Docker Hub:")
	fmt.Println("   docker login")
	fmt.Println("   Username: <your-dockerhub-username>")
	fmt.Println("   Password: <your-dockerhub-password-or-token>")
	fmt.Println()
	fmt.Println("2. For better security, use Access Tokens:")
	fmt.Println("   • Go to: https://hub.docker.com/settings/security")
	fmt.Println("   • Click 'New Access Token'")
	fmt.Println("   • Use the token as your password")
}

// showGenericGuide shows generic registry authentication guide
func showGenericGuide(registry string) {
	fmt.Println()
	fmt.Printf("🔐 Generic Registry Authentication for %s:\n", registry)
	fmt.Println()
	fmt.Printf("1. Login to the registry:\n")
	fmt.Printf("   docker login %s\n", registry)
	fmt.Println("   Username: <your-username>")
	fmt.Println("   Password: <your-password-or-token>")
	fmt.Println()
	fmt.Println("2. Check with your registry provider for:")
	fmt.Println("   • Correct authentication method")
	fmt.Println("   • Required permissions/scopes")
	fmt.Println("   • Token creation process")
}

// openRegistryDocs attempts to open registry documentation
func openRegistryDocs(regError *RegistryError) error {
	var url string

	switch regError.ErrorType {
	case "gitlab_auth":
		url = "https://docs.gitlab.com/ee/user/packages/container_registry/"
	case "github_auth":
		url = "https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry"
	case "dockerhub_auth":
		url = "https://docs.docker.com/docker-hub/"
	default:
		fmt.Println("🌐 Please check your registry provider's documentation for authentication instructions.")
		return fmt.Errorf("registry authentication required")
	}

	fmt.Printf("🌐 Opening documentation: %s\n", url)

	// Try to open the URL (macOS/Linux)
	cmd := exec.Command("open", url)
	if err := cmd.Run(); err != nil {
		// Try Linux approach
		cmd = exec.Command("xdg-open", url)
		if err := cmd.Run(); err != nil {
			fmt.Printf("💻 Please manually open: %s\n", url)
		}
	}

	return fmt.Errorf("please follow the documentation and authenticate")
}

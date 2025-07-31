package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	Version = "dev"
)

var (
	primaryColor   = lipgloss.Color("#00ADD8")
	mutedColor     = lipgloss.Color("#6B7280")
	accentColor    = lipgloss.Color("#F59E0B")
	successColor   = lipgloss.Color("#10B981")
	errorColor     = lipgloss.Color("#EF4444")
	highlightColor = lipgloss.Color("#8B5CF6")

	headerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	cmdStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Background(lipgloss.Color("#2A1A00")).
			Padding(0, 1).
			Bold(true)

	statusOkStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	statusErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	tipStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)

type WelcomeConfig struct {
	ShowStatus bool
	ShowTip    bool
	ShowTime   bool
	Compact    bool
}

func DefaultWelcomeConfig() *WelcomeConfig {
	return &WelcomeConfig{
		ShowStatus: true,
		ShowTip:    true,
		ShowTime:   true,
		Compact:    false,
	}
}

func ProjectInfo() {
	DisplayWelcome(DefaultWelcomeConfig())
}

func ProjectInfoCompact() {
	config := DefaultWelcomeConfig()
	config.Compact = true
	config.ShowTip = false
	config.ShowTime = false
	DisplayWelcome(config)
}

func DisplayWelcome(config *WelcomeConfig) {
	if config.Compact {
		displayCompact(config)
		return
	}

	displayWelcome(config)
}

func displayCompact(config *WelcomeConfig) {
	greeting := getSmartGreeting()
	
	header := headerStyle.Render("🐳 Dockyard")
	info := subtitleStyle.Render(fmt.Sprintf("%s • v%s", greeting, Version))
	
	fmt.Printf("%s %s\n", header, info)
	
	if config.ShowStatus {
		status, _ := GetDockerStatus()
		if status == "🟢" {
			fmt.Printf("   %s\n", statusOkStyle.Render("✓ Ready"))
		} else {
			fmt.Printf("   %s\n", statusErrorStyle.Render("⚠ Docker needed"))
		}
	}
}

func displayWelcome(config *WelcomeConfig) {
	greeting := getSmartGreeting()
	
	title := headerStyle.Render("🐳 Dockyard")
	subtitle := subtitleStyle.Render(fmt.Sprintf("%s • Docker Project Manager", greeting))
	version := lipgloss.NewStyle().Foreground(mutedColor).Render("v" + Version)
	
	fmt.Printf("%s\n%s %s\n", title, subtitle, version)
	
	if config.ShowTime && isDevMode() {
		timestamp := lipgloss.NewStyle().
			Foreground(mutedColor).
			Faint(true).
			Render(fmt.Sprintf("Started at %s", time.Now().Format("15:04:05")))
		fmt.Printf("%s\n", timestamp)
	}
	
	fmt.Println()

	if config.ShowStatus {
		displayEnhancedStatus()
	}

	if config.ShowTip {
		displaySmartTip()
	}
	
	if !config.Compact {
		separator := separatorStyle.Render(strings.Repeat("─", 40))
		fmt.Printf("\n%s\n", separator)
	}
}

func displayEnhancedStatus() {
	status, message := GetDockerStatus()
	
	var statusText, statusEmoji string
	if status == "🟢" {
		statusEmoji = "🟢"
		statusText = statusOkStyle.Render("Docker running")
	} else {
		statusEmoji = "🔴"
		statusText = statusErrorStyle.Render("Docker stopped")
	}
	
	context := getStatusContext()
	
	fmt.Printf("Status   %s %s", statusEmoji, statusText)
	if message != "" {
		fmt.Printf(" • %s", lipgloss.NewStyle().Foreground(mutedColor).Render(message))
	}
	fmt.Println()
	
	if context != "" {
		fmt.Printf("         %s\n", lipgloss.NewStyle().Foreground(mutedColor).Faint(true).Render(context))
	}
}

func displaySmartTip() {
	tips := getContextualTips()
	
	seed := time.Now().Hour()*60 + time.Now().Minute()/10
	tipIndex := seed % len(tips)
	
	fmt.Printf("\n%s %s\n", 
		tipStyle.Render("💡 Tip:"), 
		tips[tipIndex])
}

func getSmartGreeting() string {
	hour := time.Now().Hour()
	
	switch {
	case hour >= 5 && hour < 9:
		return "Good morning"
	case hour >= 9 && hour < 12:
		return "Morning"
	case hour >= 12 && hour < 14:
		return "Good afternoon"
	case hour >= 14 && hour < 17:
		return "Afternoon"
	case hour >= 17 && hour < 20:
		return "Good evening"
	case hour >= 20 && hour < 23:
		return "Evening"
	default:
		return "Late night coding"
	}
}

func getContextualTips() []string {
	baseCommands := []string{
		"Run " + cmdStyle.Render("dockyard list") + " to see your projects",
		"Use " + cmdStyle.Render("dockyard start") + " to launch everything",
		"Try " + cmdStyle.Render("dockyard health") + " for system checks",
		"Run " + cmdStyle.Render("dockyard logs <name>") + " to debug issues",
	}
	
	hour := time.Now().Hour()
	
	if hour < 9 {
		baseCommands = append(baseCommands, 
			"Perfect time to " + cmdStyle.Render("dockyard update") + " your images")
	} else if hour > 18 {
		baseCommands = append(baseCommands, 
			"Remember to " + cmdStyle.Render("dockyard stop") + " when done")
	}
	
	if isDevMode() {
		baseCommands = append(baseCommands,
			"Dev mode: " + cmdStyle.Render("dockyard watch") + " for live reload")
	}
	
	return baseCommands
}

func getStatusContext() string {
	status, _ := GetDockerStatus()
	
	if status == "🟢" {
		return "Ready for container operations"
	}
	
	return "Run Docker Desktop or try 'dockyard doctor'"
}

func GetDockerStatus() (string, string) {
	hour := time.Now().Hour()
	
	if hour >= 9 && hour <= 18 {
		return "🟢", "containers ready"
	}
	
	return "🟢", "daemon active"
}

func isDevMode() bool {
	return Version == "dev" || 
		   os.Getenv("DOCKYARD_ENV") == "dev" || 
		   os.Getenv("NODE_ENV") == "development"
}
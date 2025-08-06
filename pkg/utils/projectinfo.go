package utils

import (
	"fmt"
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

	tipStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)
)

type WelcomeConfig struct {
	ShowStatus bool
	ShowTip    bool
	ShowTime   bool
}

var defaultWelcome = &WelcomeConfig{
	ShowStatus: true,
	ShowTip:    true,
	ShowTime:   true,
}

func ProjectInfo() {
	displayWelcome(defaultWelcome)
}

func displayWelcome(config *WelcomeConfig) {
	greeting := SmartGreeting()

	title := headerStyle.Render("🐳 Dockyard")
	subtitle := subtitleStyle.Render(greeting)

	fmt.Printf("%s%s\n", title, subtitle)

	if config.ShowTime {
		timestamp := lipgloss.NewStyle().
			Foreground(mutedColor).
			Faint(true).
			Render(fmt.Sprintf("Started at %s", time.Now().Format("15:04:05")))
		fmt.Printf("%s\n", timestamp)
	}

	fmt.Println()

	if config.ShowTip {
		displaySmartTip()
	}
}

func displaySmartTip() {
	tips := getContextualTips()

	now := time.Now()
	seed := now.Hour()*60 + now.Minute()/10
	tipIndex := seed % len(tips)

	fmt.Printf("\n%s %s\n",
		tipStyle.Render("💡 Tip:"),
		tips[tipIndex])
}

var (
	cutoffs  = []int{5, 9, 12, 14, 17, 20, 24}
	messages = []string{
		"Late night coding",
		"Good morning",
		"Morning",
		"Good afternoon",
		"Afternoon",
		"Good evening",
		"Evening",
	}
)

func SmartGreeting() string {
	h := time.Now().Hour()
	for i, t := range cutoffs {
		if h < t {
			return messages[i]
		}
	}
	return messages[0]
}

func getContextualTips() []string {
	baseCommands := []string{
		"Run " + cmdStyle.Render("dockyard list") + " to see your projects",
		"Use " + cmdStyle.Render("dockyard start") + " to launch everything",
		"Try " + cmdStyle.Render("dockyard health") + " for system checks",
	}

	hour := time.Now().Hour()

	if hour < 9 {
		baseCommands = append(baseCommands,
			"Perfect time to "+cmdStyle.Render("dockyard update")+" your images")
	} else if hour > 18 {
		baseCommands = append(baseCommands,
			"Remember to "+cmdStyle.Render("dockyard stop")+" when done")
	}

	return baseCommands
}

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	primaryColor = lipgloss.Color("#00ADD8") // Go blue
	successColor = lipgloss.Color("#00D484") // Green
	errorColor   = lipgloss.Color("#ED567A") // Red
	warningColor = lipgloss.Color("#FFAB00") // Orange
	infoColor    = lipgloss.Color("#7C3AED") // Purple
	mutedColor   = lipgloss.Color("#6B7280") // Gray
	accentColor  = lipgloss.Color("#F59E0B") // Amber

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Header styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)

	headerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)

	// Status styles
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			Bold(true)

	// Content styles
	codeStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(0, 1).
			Italic(true)

	listStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			MarginBottom(1)

	instructionStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				PaddingLeft(4)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2).
			MarginBottom(1)

	highlightBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1, 2).
				MarginBottom(1)

	// Progress styles
	spinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)
)

// Render functions for different message types
func RenderTitle(text string) string {
	return titleStyle.Render(text)
}

func RenderHeader(text string) string {
	return headerStyle.Render(text)
}

func RenderSuccess(text string) string {
	return successStyle.Render("✅ " + text)
}

func RenderError(text string) string {
	return errorStyle.Render("❌ " + text)
}

func RenderWarning(text string) string {
	return warningStyle.Render("⚠️  " + text)
}

func RenderInfo(text string) string {
	return infoStyle.Render("ℹ️  " + text)
}

func RenderCode(text string) string {
	return codeStyle.Render(text)
}

func RenderBox(content string) string {
	return boxStyle.Render(content)
}

func RenderHighlightBox(content string) string {
	return highlightBoxStyle.Render(content)
}

// Render lists with proper styling
func RenderList(items []string) string {
	var styledItems []string
	for _, item := range items {
		if strings.HasPrefix(item, "   ") {
			// Instruction/sub-item
			styledItems = append(styledItems, instructionStyle.Render(item))
		} else if strings.Contains(item, ":") && (strings.Contains(item, "•") || strings.Contains(item, "Run:")) {
			// Command or instruction
			parts := strings.SplitN(item, ":", 2)
			if len(parts) == 2 {
				styled := parts[0] + ": " + codeStyle.Render(strings.TrimSpace(parts[1]))
				styledItems = append(styledItems, "     "+styled)
			} else {
				styledItems = append(styledItems, item)
			}
		} else if strings.HasPrefix(item, "🍎") || strings.HasPrefix(item, "🪟") || strings.HasPrefix(item, "🐧") {
			// Platform headers
			styledItems = append(styledItems, headerStyle.Render(item))
		} else if strings.HasPrefix(item, "📖") || strings.HasPrefix(item, "🚀") || strings.HasPrefix(item, "⚙️") {
			// Section headers
			styledItems = append(styledItems, infoStyle.Render(item))
		} else if item == "" {
			styledItems = append(styledItems, "")
		} else {
			styledItems = append(styledItems, listStyle.Render(item))
		}
	}
	return strings.Join(styledItems, "\n")
}

// Render markdown content with glamour
func RenderMarkdown(content string) string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return content // fallback to plain text
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		return content // fallback to plain text
	}

	return rendered
}

// Create formatted sections
func RenderSection(title, content string) string {
	header := headerStyle.Render(title)
	body := baseStyle.Render(content)
	return fmt.Sprintf("%s\n%s", header, body)
}

// Create progress indicator
func RenderProgress(message string, current, total int) string {
	progressBar := createProgressBar(current, total)
	return fmt.Sprintf("%s %s (%d/%d)",
		spinnerStyle.Render("⏳"),
		message,
		current,
		total,
	) + "\n" + progressBar
}

func createProgressBar(current, total int) string {
	if total == 0 {
		return ""
	}

	width := 20
	filled := int(float64(current) / float64(total) * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	percentage := int(float64(current) / float64(total) * 100)

	return fmt.Sprintf("[%s] %d%%",
		lipgloss.NewStyle().Foreground(primaryColor).Render(bar),
		percentage,
	)
}

// Platform-specific icon styling
func RenderPlatformIcon(platform string) string {
	switch platform {
	case "darwin":
		return "🍎"
	case "windows":
		return "🪟"
	case "linux":
		return "🐧"
	default:
		return "🖥️"
	}
}

// Runtime-specific icon styling
func RenderRuntimeIcon(runtime string) string {
	switch runtime {
	case "orbstack":
		return "🌌"
	case "colima":
		return "🦙"
	case "docker-desktop":
		return "🐳"
	case "podman":
		return "🦭"
	default:
		return "📦"
	}
}

// Status indicator styles
func RenderStatusCheck(message string) string {
	return infoStyle.Render(fmt.Sprintf("🔍 %s", message))
}

func RenderQuickStatus(message string, success bool) string {
	if success {
		return fmt.Sprintf("%s %s",
			successStyle.Render("✅"),
			message)
	}
	return fmt.Sprintf("%s %s",
		errorStyle.Render("❌"),
		message)
}

// Inline status rendering for brief checks
func RenderInlineStatus(message string) string {
	return infoStyle.Render(message)
}

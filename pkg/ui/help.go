package ui

import "fmt"

const platformHelpMarkdown = `
# 🐳 Container Runtime Setup Guide

## Supported Platforms

### 🍎 macOS
Choose from these excellent container runtimes:

- **OrbStack** ⭐ *Recommended*
  - Fast, lightweight, native macOS experience
  - Install: [orbstack.dev](https://orbstack.dev)
  - Auto-start: Settings → General → "Start on login"

- **Docker Desktop**
  - Official Docker runtime
  - Install: [docker.com](https://docker.com/products/docker-desktop)
  - Auto-start: Settings → General → "Start when you log in"

- **Colima**
  - Lightweight VM-based solution
  - Install: ` + "`brew install colima docker`" + `
  - Auto-start: ` + "`brew services start colima`" + `

- **Podman**
  - Daemonless container engine
  - Install: ` + "`brew install podman`" + `
  - Start: ` + "`podman machine start`" + `

### 🪟 Windows
- **Docker Desktop**: Full-featured Windows container runtime
- **Podman Desktop**: Alternative container management

### 🐧 Linux
- **Docker Engine**: Native Linux container runtime
- **Podman**: Red Hat's container engine

## 🚀 Quick Commands

` + "```bash" + `
# Check status
dockyard health

# Start containers
dockyard start

# View logs  
dockyard logs <project>
` + "```" + `

> 💡 **Tip**: Run ` + "`dockyard --help`" + ` for all available commands!
`

func RenderPlatformHelp() string {
	return RenderMarkdown(platformHelpMarkdown)
}

func RenderRuntimeStatus(runtime, status string) string {
	icon := RenderRuntimeIcon(runtime)
	switch status {
	case "running":
		return RenderSuccess(fmt.Sprintf("%s %s is running", icon, runtime))
	case "stopped":
		return RenderError(fmt.Sprintf("%s %s is stopped", icon, runtime))
	case "starting":
		return RenderInfo(fmt.Sprintf("%s %s is starting...", icon, runtime))
	default:
		return RenderWarning(fmt.Sprintf("%s %s status unknown", icon, runtime))
	}
}

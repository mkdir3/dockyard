package cmd

import (
	"dockyard/pkg/docker"
	"dockyard/pkg/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	follow bool
)

var logsCmd = &cobra.Command{
	Use:   "logs [project] [service...]",
	Short: "View logs for services in a project",
	Long:  `Display logs for specific services within a Docker project. If no services specified, shows logs for all services.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		var targetServices []string

		if len(args) > 1 {
			targetServices = args[1:]
		}

		projectPath, ok := docker.Projects[projectName]
		if !ok {
			fmt.Printf("Unknown project: %s\n", projectName)
			return
		}

		projectDir, err := utils.ResolveHomeDir(projectPath)
		if err != nil {
			fmt.Printf("Failed to resolve home directory in %s: %v\n", projectPath, err)
			return
		}

		cm, err := docker.NewComposeManager()
		if err != nil {
			fmt.Printf("Failed to create compose manager: %v\n", err)
			return
		}
		defer cm.Close()

		if err := cm.ViewLogs(projectDir, targetServices, follow); err != nil {
			fmt.Printf("Failed to view logs: %v\n", err)
			return
		}
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	rootCmd.AddCommand(logsCmd)
}

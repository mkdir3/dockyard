package cmd

import (
	"dockyard/pkg/docker"
	"dockyard/pkg/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	removeOrphans bool
	detached      bool
)

var startCmd = &cobra.Command{
	Use:   "start [project]",
	Short: "Start a Docker project",
	Long:  `Start all Docker containers of a project using Docker Compose`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
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
		defer func(cm *docker.ComposeManager) {
			err := cm.Close()
			if err != nil {
				fmt.Printf("Failed to close compose manager: %v\n", err)
			} else {
				fmt.Println("✅ Compose manager connection closed")
			}
		}(cm)

		err = cm.StartProject(projectDir, detached, removeOrphans)
		if err != nil {
			fmt.Printf("Failed to start project %s: %v\n", projectName, err)
			return
		}

		fmt.Printf("✅ Project %s started successfully!\n", projectName)
	},
}

func init() {
	startCmd.Flags().BoolVar(&removeOrphans, "remove-orphans", true, "Remove containers for services not defined in the Compose file")
	startCmd.Flags().BoolVarP(&detached, "detach", "d", true, "Detached mode: Run containers in the background")
	rootCmd.AddCommand(startCmd)
}

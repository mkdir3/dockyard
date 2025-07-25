package cmd

import (
	"dockyard/pkg/docker"
	"dockyard/pkg/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [project]",
	Short: "Pull images for a Docker project",
	Long:  `Pull service images for a Docker project`,
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

		err = cm.PullImages(projectDir)
		if err != nil {
			fmt.Printf("Failed to pull images for project %s: %v\n", projectName, err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

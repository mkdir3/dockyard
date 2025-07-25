package cmd

import (
	"dockyard/pkg/docker"
	"dockyard/pkg/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	noCache bool
)

var buildCmd = &cobra.Command{
	Use:   "build [project]",
	Short: "Build images for a Docker project",
	Long:  `Build or rebuild services in a Docker project`,
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

		err = cm.BuildImages(projectDir, noCache)
		if err != nil {
			fmt.Printf("Failed to build project %s: %v\n", projectName, err)
			return
		}
	},
}

func init() {
	buildCmd.Flags().BoolVar(&noCache, "no-cache", false, "Do not use cache when building the image")
	rootCmd.AddCommand(buildCmd)
}

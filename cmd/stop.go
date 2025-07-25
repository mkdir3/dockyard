package cmd

import (
	"dockyard/pkg/docker"
	"dockyard/pkg/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	removeVolumes bool
	removeImages  bool
)

var stopCmd = &cobra.Command{
	Use:   "stop [project]",
	Short: "Stop a Docker project",
	Long:  `Stop a Docker project by its name`,
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

		err = cm.StopProject(projectDir, removeVolumes, removeImages)
		if err != nil {
			fmt.Printf("Failed to stop project %s: %v\n", projectName, err)
			return
		}
	},
}

func init() {
	stopCmd.Flags().BoolVarP(&removeVolumes, "volumes", "v", false, "Remove named volumes declared in the volumes section and anonymous volumes")
	stopCmd.Flags().BoolVar(&removeImages, "rmi", false, "Remove images used by services")
	rootCmd.AddCommand(stopCmd)
}

package cmd

import (
	"dockyard/pkg/docker"
	"dockyard/pkg/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Docker projects",
	Long:  `List all Docker projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Projects:")
		sortedProjectNames := docker.GetSortedProjectNames()
		for _, projectName := range sortedProjectNames {
			projectPath := docker.Projects[projectName]
			projectDir, err := utils.ResolveHomeDir(projectPath)
			if err != nil {
				fmt.Printf("Failed to resolve home directory in %s: %v\n", projectPath, err)
				continue
			}
			composeFilePath, err := utils.GetComposeFilePath(projectDir)
			if err != nil {
				fmt.Printf("Failed to find docker-compose file in %s: %v\n", projectDir, err)
				continue
			}
			fmt.Printf("- %s (%s)\n", projectName, composeFilePath)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

package cmd

import (
	"dockyard/pkg/docker"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Manage Docker projects",
	Long:  `Add or remove Docker projects paths.`,
	Run: func(cmd *cobra.Command, args []string) {
		var action string
		actionPrompt := &survey.Select{
			Message: "What do you want to do?",
			Options: []string{"Add", "Remove"},
		}
		err := survey.AskOne(actionPrompt, &action)
		if err != nil {
			return
		}

		switch action {
		case "Add":
			err := docker.AddProject()
			if err != nil {
				return
			}

		case "Remove":
			err := docker.RemoveProject()
			if err != nil {
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(manageCmd)
}

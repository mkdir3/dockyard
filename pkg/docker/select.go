package docker

import (
	"github.com/AlecAivazis/survey/v2"
)

func SelectProjects() ([]string, error) {
	projectNames := GetSortedProjectNames()

	var selectedProjects []string
	prompt := &survey.MultiSelect{
		Message: "Which projects do you want to start?",
		Options: projectNames,
	}
	err := survey.AskOne(prompt, &selectedProjects)
	if err != nil {
		return nil, err
	}

	return selectedProjects, nil
}

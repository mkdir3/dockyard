package main

import (
	"dockyard/cmd"
	"dockyard/pkg/utils"
)

func main() {
	utils.ProjectInfo()
	cmd.Execute()
}

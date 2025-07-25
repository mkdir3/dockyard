package utils

import (
	"fmt"
)

const (
	Reset  = "\033[0m"
	Purple = "\033[95m"
	Cyan   = "\033[96m"
	White  = "\033[97m"
)

func ProjectInfo() {
	fmt.Printf(
		"%s🚀 Welcome to Dockyard! Your best local project manager! 🐳%s\n"+
			"%s📍 Made by: tenderpanini%s\n",
		Purple, Reset,
		Cyan, Reset,
	)
}

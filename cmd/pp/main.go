package main

import (
	"github.com/from68/pp-cli/commands"
)

var Version = "dev"

func main() {
	commands.SetVersion(Version)
	commands.Execute()
}

package main

import (
	"mayday/core/commands"
)

func main() {
	commands.Parse(
		new(commands.RunCommand),
		new(commands.PushCommand),
		new(commands.PullCommand),
	)
}

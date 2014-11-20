package main

import (
	"mayday/core/commands"
)

func main() {
	commands.Parse(
		new(commands.RunCommand),
		new(commands.UpdateCommand),
		new(commands.PullCommand),
		new(commands.ShowCommand),
		new(commands.CreateCommand),
	)
}

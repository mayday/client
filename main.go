package main

import (
	"mayday/core"
	"mayday/core/commands"
)

func main() {

	env := core.DefaultEnvironment{}

	commands.Parse(env,
		new(commands.RunCommand),
		new(commands.UpdateCommand),
		new(commands.PullCommand),
		new(commands.ShowCommand),
		new(commands.CreateCommand),
		new(commands.ServerCommand),
	)
}

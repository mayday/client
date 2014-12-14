package commands

import (
	"flag"
	"mayday/core"
	"mayday/server"
)

type ServerCommand struct {
	port    *int
	bind    *string
	storage *string
}

func (cmd *ServerCommand) Name() string {
	return "server"
}

func (cmd *ServerCommand) Description() string {
	return "Start a mayday server."
}

func (cmd *ServerCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.storage = fs.String("storage", "", "Storage path for case report files")
	cmd.port = fs.Int("port", server.DefaultPort, "Port to bind the mayday server")
	cmd.bind = fs.String("bind", "0.0.0.0", "Address to bind the mayday server")
}

func (cmd *ServerCommand) Run(env core.Environment) {
	server.Start(env, *cmd.bind, *cmd.port, *cmd.storage)
}

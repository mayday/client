package commands

import (
	"flag"
	"mayday/core"
)

type PullCommand struct {
	token  *string
	uuid   *string
	server *string
}

func (cmd *PullCommand) Name() string {
	return "pull"
}

func (cmd *PullCommand) Description() string {
	return "Get the current reports files for a case."
}

func (cmd *PullCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.uuid = fs.String("uuid", "", "PGP KeyID to sign the new configuration")
	cmd.token = fs.String("token", "", "PGP KeyID to sign the new configuration")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
}

func (cmd *PullCommand) Run() {
	// real stuff
}

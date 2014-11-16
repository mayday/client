package commands

import (
	"flag"
	"mayday/core"
)

type ShowCommand struct {
	token  *string
	uuid   *string
	server *string
}

func (cmd *ShowCommand) Name() string {
	return "show"
}

func (cmd *ShowCommand) Description() string {
	return "Get the current reports files for a case."
}

func (cmd *ShowCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.uuid = fs.String("uuid", "", "PGP KeyID to sign the new configuration")
	cmd.token = fs.String("token", "", "PGP KeyID to sign the new configuration")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
}

func (cmd *ShowCommand) Run() {
	// real stuff
}

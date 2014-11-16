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
	return "Show the current configuration and files for a specific case."
}

func (cmd *ShowCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.uuid = fs.String("uuid", "", "Case UUID")
	cmd.token = fs.String("auth", "", "Case authentication token")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
}

func (cmd *ShowCommand) Run() {
	// real stuff
}

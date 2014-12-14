package commands

import (
	"flag"
	"mayday/core"
)

type UpdateCommand struct {
	pgpkeyid *string
	server   *string
	uuid     *string
}

func (cmd *UpdateCommand) Name() string {
	return "update"
}

func (cmd *UpdateCommand) Description() string {
	return "Update the configuration file for a case"
}

func (cmd *UpdateCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.uuid = fs.String("uuid", "", "PGP KeyID to sign the new configuration")
	cmd.pgpkeyid = fs.String("keyid", "", "PGP KeyID to sign the new configuration")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
}

func (cmd *UpdateCommand) Run(env core.Environment) {
	// real stuff
}

package commands

import (
	"flag"
	"mayday/core"
)

type PushCommand struct {
	pgpkeyid *string
	server   *string
	uuid     *string
}

func (cmd *PushCommand) Name() string {
	return "push"
}

func (cmd *PushCommand) Description() string {
	return "Create a new case and upload the configuration file"
}

func (cmd *PushCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.uuid = fs.String("uuid", "", "PGP KeyID to sign the new configuration")
	cmd.pgpkeyid = fs.String("pgp-keyid", "", "PGP KeyID to sign the new configuration")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
}

func (cmd *PushCommand) Run() {
	// real stuff
}

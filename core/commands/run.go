package commands

import (
	"flag"
	"fmt"
	"mayday/core"
	"os"
)

type RunCommand struct {
	pgp    *bool
	uuid   *string
	server *string
	auth   *string
	upload *bool
}

func (cmd *RunCommand) Name() string {
	return "run"
}

func (cmd *RunCommand) Description() string {
	return "Run a specific case and generate the reports."
}

func (cmd *RunCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.pgp = fs.Bool("no-pgp", false, "Disable pgp signature validation")
	cmd.upload = fs.Bool("no-upload", false, "Don't upload generated reports")
	cmd.uuid = fs.String("uuid", "", "Mayday server address")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
	cmd.auth = fs.String("auth", "", "PGP KeyID to sign the new configuration")
}

func (cmd *RunCommand) Run() {
	if *cmd.uuid == "" {
		fmt.Println("Please specify a case UUID --uuid UUID\n")
		os.Exit(1)
	}

	if *cmd.auth == "" {
		fmt.Println("Please specify the case auth token --auth TOKEN\n")
		os.Exit(1)
	}
	mayday, err := core.NewClient(*cmd.server, *cmd.uuid, *cmd.auth)
	if err != nil {
		fmt.Println(err)
	}

	err = mayday.Run(!*cmd.pgp, *cmd.upload)
	if err != nil {
		fmt.Println(err)
	}
}

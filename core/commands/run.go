package commands

import (
	"flag"
	"fmt"
	"mayday/core"
	"os"
)

type RunCommand struct {
	id      *string
	pgp     *bool
	dryRun  *bool
	server  *string
	timeout *int
	token   *string
	upload  *bool
}

func (cmd *RunCommand) Name() string {
	return "run"
}

func (cmd *RunCommand) Description() string {
	return "Run a specific case and generate the reports."
}

func (cmd *RunCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.id = fs.String("id", "", "Case id")
	cmd.pgp = fs.Bool("pgp", true, "Enable pgp signature validation")
	cmd.dryRun = fs.Bool("dry-run", true, "Enable pgp signature validation")
	cmd.server = fs.String("server", "", "Mayday server address")
	cmd.timeout = fs.Int("timeout", 0, "Default timeout for commands")
	cmd.token = fs.String("token", "", "Authentication token for the case")
	cmd.upload = fs.Bool("upload", true, "Upload the generated reports to the server")
}

func (cmd *RunCommand) Run() {
	if *cmd.id == "" {
		fmt.Println("Please specify a Case Id --id\n")
		os.Exit(1)
	}

	mayday, err := core.NewClient(*cmd.server, *cmd.id, *cmd.token)
	if err != nil {
		fmt.Println(err)
	}

	err = mayday.Run(*cmd.pgp, *cmd.upload, *cmd.timeout, *cmd.dryRun)
	if err != nil {
		fmt.Println(err)
	}
}

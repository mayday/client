package commands

import (
	"flag"
	"fmt"
	"mayday/core"
	"os"
)

type CreateCommand struct {
	pgp         *bool
	server      *string
	private     *bool
	config      *string
	description *string
	keyid       *string
}

func (cmd *CreateCommand) Name() string {
	return "create"
}

func (cmd *CreateCommand) Description() string {
	return "Create a specific case and generate the reports."
}

func (cmd *CreateCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.pgp = fs.Bool("pgp", true, "Disable pgp signature validation")
	cmd.keyid = fs.String("keyid", "", "GPG Key ID to use")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
	cmd.private = fs.Bool("private", false, "Disable pgp signature validation")
	cmd.description = fs.String("description", "", "Mayday server address")
	cmd.config = fs.String("config", "", "Configuration file for case")
}

func (cmd *CreateCommand) Run(env core.Environment) {
	mayday, err := core.NewClient(env, *cmd.server, "", "")

	if err != nil {
		fmt.Println(err)
	}

	if *cmd.config == "" {
		fmt.Println("Please specify --config path")
		os.Exit(-1)
	}

	new_case, err := mayday.Create(*cmd.config, *cmd.description, *cmd.private, *cmd.pgp, *cmd.keyid)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	fmt.Println(new_case)

}

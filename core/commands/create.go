package commands

import (
	"flag"
	"fmt"
	"mayday/core"
)

type CreateCommand struct {
	pgp       *bool
	uuid      *string
	server    *string
	auth      *string
	upload    *bool
	dryCreate *bool
	timeout   *int
}

func (cmd *CreateCommand) Name() string {
	return "create"
}

func (cmd *CreateCommand) Description() string {
	return "Create a specific case and generate the reports."
}

func (cmd *CreateCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.pgp = fs.Bool("no-pgp", false, "Disable pgp signature validation")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
}

func (cmd *CreateCommand) Run() {
	mayday, err := core.NewClient(*cmd.server, "", "")

	if err != nil {
		fmt.Println(err)
	}

	itf, err := mayday.Create()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(itf)
}

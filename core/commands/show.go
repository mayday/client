package commands

import (
	"flag"
	"fmt"
	"mayday/core"
	"os"
)

type ShowCommand struct {
	auth   *string
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
	cmd.auth = fs.String("auth", "", "Case authentication token")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
}

func (cmd *ShowCommand) Run() {
	if *cmd.uuid == "" || *cmd.auth == "" {
		flag.Usage()
		os.Exit(1)
	}

	mayday, err := core.NewClient(*cmd.server, *cmd.uuid, *cmd.auth)

	if err != nil {
		fmt.Println(err)
	}

	uuid, config, err := mayday.Show()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Current configuration for case id: %s\n\n\n", uuid)
	fmt.Printf("%s\n", config)

	// real stuff
}

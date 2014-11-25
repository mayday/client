package commands

import (
	"flag"
	"fmt"
	"io/ioutil"
	"mayday/core"
	"os"
	"path"
)

type PullCommand struct {
	token  *string
	id     *string
	server *string
	all    *bool
	fileId *string
	to     *string
}

func (cmd *PullCommand) Name() string {
	return "pull"
}

func (cmd *PullCommand) Description() string {
	return "Get the current reports files for a case."
}

func (cmd *PullCommand) DefineFlags(fs *flag.FlagSet) {
	cmd.id = fs.String("id", "", "Case ID to pull from")
	cmd.all = fs.Bool("all", true, "Pull all files from the case")
	cmd.fileId = fs.String("file-id", "", "File Id to retrieve")
	cmd.to = fs.String("to", "", "Path to store the retrieved files")
	cmd.token = fs.String("token", "", "PGP KeyID to sign the new configuration")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
}

func (cmd *PullCommand) Run() {
	if *cmd.id == "" {
		fmt.Println("Please specify a valid Case Id --id\n")
		os.Exit(1)
	}

	mayday, err := core.NewClient(*cmd.server, *cmd.id, *cmd.token)
	if err != nil {
		fmt.Println(err)
	}

	var files map[string]string

	if *cmd.fileId != "" {
		files, err = mayday.Pull(*cmd.fileId)
		if err != nil {
			fmt.Println(err)
		}

	} else if !*cmd.all {
		fmt.Println("Please specify a --file-id or --all")
		os.Exit(1)
	} else {
		files, err = mayday.PullAll()
		if err != nil {
			fmt.Println(err)
		}
	}

	for name, content := range files {
		filename := ""
		if *cmd.to == "" {
			base, err := core.GetDefaultDirectory()
			if err != nil {
				fmt.Println(err)
			}

			base = path.Join(base, "pull", *cmd.id)
			if _, err := os.Stat(base); os.IsNotExist(err) {
				err := os.MkdirAll(base, 0700)
				if err != nil {
					fmt.Println(err)
				}
			}

			filename = path.Join(base, name)
		} else {
			filename = path.Join(*cmd.to, name)
		}

		err := ioutil.WriteFile(filename, []byte(content), 0700)
		if err != nil {
			fmt.Println("Error writing file")
		} else {
			fmt.Printf("Pulled report on path: %s\n", filename)
		}
	}

	// real stuff
}

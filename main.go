package main

import (
	"flag"
	"fmt"
	"mayday/core"
	"os"
)

type subCommand interface {
	Name() string
	DefineFlags(*flag.FlagSet)
	Run()
}

type subCommandParser struct {
	cmd subCommand
	fs  *flag.FlagSet
}

func Parse(commands ...subCommand) {
	scp := make(map[string]*subCommandParser, len(commands))
	for _, cmd := range commands {
		name := cmd.Name()
		scp[name] = &subCommandParser{cmd, flag.NewFlagSet(name, flag.ExitOnError)}
		cmd.DefineFlags(scp[name].fs)
	}

	oldUsage := flag.Usage
	flag.Usage = func() {
		oldUsage()
		fmt.Println(scp)
		for name, sc := range scp {
			fmt.Fprintf(os.Stderr, "\n# %s %s\n", os.Args[0], name)
			sc.fs.PrintDefaults()
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	cmdname := flag.Arg(0)
	if sc, ok := scp[cmdname]; ok {
		sc.fs.Parse(flag.Args()[1:])
		sc.cmd.Run()
	} else {
		fmt.Fprintf(os.Stderr, "error: %s is not a valid command", cmdname)
		flag.Usage()
		os.Exit(1)
	}
}

type push struct {
	pgpkeyid *string
	server   *string
	uuid     *string
}

func (cmd *push) Name() string {
	return "push"
}

func (cmd *push) DefineFlags(fs *flag.FlagSet) {
	cmd.uuid = fs.String("uuid", "", "PGP KeyID to sign the new configuration")
	cmd.pgpkeyid = fs.String("pgp-keyid", "", "PGP KeyID to sign the new configuration")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
}

func (cmd *push) Run() {
	// real stuff
}

type pull struct {
	token  *string
	uuid   *string
	server *string
}

func (cmd *pull) Name() string {
	return "pull"
}

func (cmd *pull) DefineFlags(fs *flag.FlagSet) {
	cmd.uuid = fs.String("uuid", "", "PGP KeyID to sign the new configuration")
	cmd.token = fs.String("token", "", "PGP KeyID to sign the new configuration")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
}

func (cmd *pull) Run() {
	// real stuff
}

type run struct {
	pgp    *bool
	uuid   *string
	server *string
	auth   *string
	upload *bool
}

func (cmd *run) Name() string {
	return "run"
}

func (cmd *run) DefineFlags(fs *flag.FlagSet) {
	cmd.pgp = fs.Bool("no-pgp", false, "Disable pgp signature validation")
	cmd.upload = fs.Bool("no-upload", false, "Don't upload generated reports")
	cmd.uuid = fs.String("uuid", "", "Mayday server address")
	cmd.server = fs.String("server", core.DefaultAPIBaseURL, "Mayday server address")
	cmd.auth = fs.String("auth", "", "PGP KeyID to sign the new configuration")
}

func (cmd *run) Run() {
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

	err = mayday.Run(*cmd.pgp, *cmd.upload)
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println("Viva la vida")
	// real stuff

}

func main() {
	Parse(new(run), new(push), new(pull))
}

// func main() {

// 	subcommands := make(map[string]*flag.FlagSet, 3)

// 	flag.Bool("run", false, "test")
// 	flag.Bool("push", false, "test")
// 	flag.Bool("pull", false, "test")

// 	flag.Parse()

// 	if flag.NArg() < 1 {
// 		flag.Usage()
// 		os.Exit(1)
// 	}

// 	cmdName := flag.Arg(0)

// 	config, err := core.NewConfig("./config.yaml", "./freyes.config.yaml.sig")

// 	if err != nil {
// 		fmt.Printf("Error %s\n", err)
// 		return
// 	}

// 	mayday, err := core.NewClient(config)
// 	mayday.Run()

// }

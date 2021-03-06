package commands

import (
	"flag"
	"fmt"
	"mayday/core"
	"os"
)

type subCommand interface {
	Name() string
	Description() string
	DefineFlags(*flag.FlagSet)
	Run(env core.Environment)
}

type subCommandParser struct {
	cmd subCommand
	fs  *flag.FlagSet
}

func Parse(env core.Environment, commands ...subCommand) {
	scp := make(map[string]*subCommandParser, len(commands))
	for _, cmd := range commands {
		name := cmd.Name()
		scp[name] = &subCommandParser{cmd, flag.NewFlagSet(name, flag.ExitOnError)}
		cmd.DefineFlags(scp[name].fs)
	}

	oldUsage := flag.Usage
	flag.Usage = func() {
		oldUsage()
		for name, sc := range scp {
			fmt.Fprintf(os.Stderr, "\n%s %s\n%s\n", os.Args[0], name, sc.cmd.Description())
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
		sc.cmd.Run(env)
	} else {
		fmt.Fprintf(os.Stderr, "error: %s is not a valid command", cmdname)
		flag.Usage()
		os.Exit(1)
	}
}

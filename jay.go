package main

import (
	"fmt"
	"log"
	"os"

	"github.com/blue-jay/jay/command"
	"github.com/blue-jay/jay/console"
	"github.com/blue-jay/jay/env"
	"github.com/blue-jay/jay/find"
	"github.com/blue-jay/jay/generate"
	"github.com/blue-jay/jay/migrate"
	"github.com/blue-jay/jay/replace"
)

// Commands lists the available commands.
// The order here is the order in which they are printed by 'jay help'.
var commands = []*command.Info{
	find.Cmd,
	replace.Cmd,
	migrate.Cmd,
	generate.Cmd,
	env.Cmd,
	console.Cmd,
}

func init() {
	// Verbose logging with file name and line number
	log.SetFlags(log.Lshortfile)
}

func main() {
	// Set the commands and templates
	command.Config(commands, documentationTemplate, helpTemplate, usageTemplate)

	if len(os.Args[1:]) < 1 {
		command.Usage()
		os.Exit(2)
	}

	if os.Args[1] == "help" {
		command.Help(os.Args[2:])
		return
	}

	args := os.Args[1:]
	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Runnable() {
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Run(cmd, args[1:])
			os.Exit(0)
			return
		}
	}
	fmt.Fprintf(os.Stderr, "jay: unknown subcommand %q\nRun 'jay help' for usage.\n", args[0])
	os.Exit(2)
}

var usageTemplate = `Jay is a tool for manipulating code.

Usage:

	jay command [arguments]
	
The commands are:
{{range .}}{{if .Runnable}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "jay help [command]" for more information about a command.
`

var helpTemplate = `{{if .Runnable}}usage: jay {{.UsageLine}}

{{end}}{{.Long | trim}}
`

var documentationTemplate = `/*
{{range .}}{{if .Short}}{{.Short | capitalize}}

{{end}}{{if .Runnable}}Usage:

	jay {{.UsageLine}}

{{end}}{{.Long | trim}}


{{end}}*/
package main
`

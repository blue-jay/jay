package console

import (
	"fmt"

	"github.com/blue-jay/jay/command"
	"github.com/blue-jay/jay/console/goconsole"
)

var Cmd = &command.Info{
	Run:       runConsole,
	UsageLine: "console",
	Short:     "load a terminal that doesn't really do anything",
	Long: `
Console will load a terminal.
`,
}

func runConsole(cmd *command.Info, args []string) {
	con := goconsole.New()
	con.Title = "*** Jay Console ***\n"
	con.Prompt = "> "
	con.NotFound = "Command not found: "
	con.NewLine = "\n"
	con.Add("hello", "Prints: world", func(typed string) {
		fmt.Print("world")
	})
	con.Start()
}

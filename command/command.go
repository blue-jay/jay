// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

var (
	documentationTemplate string
	helpTemplate          string
	usageTemplate         string
	commands              []*Info
)

// Config sets the templates
func Config(com []*Info, doc, help, usage string) {
	commands = com
	documentationTemplate = doc
	helpTemplate = help
	usageTemplate = usage
}

// Info is an implementation of a jay command.
type Info struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Info, args []string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'jay help' output.
	Short string

	// Long is the long message shown in the 'jay help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

// Name returns the command's name: the first word in the usage line.
func (c *Info) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Info) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

// Runnable reports whether the command can be run; otherwise
// it is a documentation pseudo-command such as importpath.
func (c *Info) Runnable() bool {
	return c.Run != nil
}

// Usage outputs a list of the commands
func Usage() {
	printUsage(os.Stderr)
}

func printUsage(w io.Writer) {
	bw := bufio.NewWriter(w)
	tmpl(bw, usageTemplate, commands)
	bw.Flush()
}

// tmpl executes the given template text on data, writing the result to w.
func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace, "capitalize": capitalize})
	template.Must(t.Parse(text))
	err := t.Execute(w, data)
	if err != nil {
		log.Fatalln(err)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}

// Help implements the 'help' command.
func Help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		// not exit 2: succeeded at 'jay help'.
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: jay help command\n\nToo many arguments given.\n")
		os.Exit(2) // failed at 'jay help'
	}

	arg := args[0]

	// 'jay help documentation' generates doc.go.
	if arg == "documentation" {
		buf := new(bytes.Buffer)
		printUsage(buf)
		usage := &Info{Long: buf.String()}
		tmpl(os.Stdout, documentationTemplate, append([]*Info{usage}, commands...))
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			// not exit 2: succeeded at 'jay help cmd'.
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run 'jay help'.\n", arg)
	os.Exit(2) // failed at 'jay help cmd'
}

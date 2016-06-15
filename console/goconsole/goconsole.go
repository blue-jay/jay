/*
Package goconsole provides simple command line functionality.

Console output after typing a few keywords:

	*** Jay Console ***
	> help
	Available commands:
	exit - Exit the console
	hello - Prints: world
	help - Show a list of available commands

	> hello
	world

	>
*/
package goconsole

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Structure of a command
type command struct {
	name        string
	description string
	function    func(string)
}

// Console configurations
type Console struct {
	// Controls the console loop
	Active bool

	// Displayed at the beginning of a line. Defaults to no linebreak. Possible value is "\n".
	Prompt string

	// Displayed when a command is not found.
	NotFound string

	// Displayed when first opened on the top line.
	Title string

	// Displayed between each line
	NewLine string

	// Map of all the callable commands
	commands map[string]command
}

// The New function creates an instance of the console type.
func New() *Console {
	con := &Console{}
	con.commands = make(map[string]command)
	con.Active = false
	con.Prompt = ""
	con.NotFound = "Not found:"
	con.Title = ""
	con.NewLine = ""

	// Load the default commands
	con.Add("exit", "Exit the console", func(typed string) {
		con.Active = false
		fmt.Print("Goodbye.")
	})

	con.Add("help", "Show a list of available commands", func(typed string) {
		// Sort by keywords
		keys := make([]string, 0)
		for key := range con.commands {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Output the commands
		fmt.Println("Available commands:")
		for i, val := range keys {
			if i == len(keys)-1 {
				fmt.Print(con.commands[val].name, " - ", con.commands[val].description)
			} else {
				fmt.Println(con.commands[val].name, "-", con.commands[val].description)
			}
		}
	})

	return con
}

// *****************************************************************************
// Core
// *****************************************************************************

// The Start function starts the console loop where the user is prompted for keywords and then runs the associated functions.
func (con *Console) Start() {
	fmt.Print(con.Title)

	// Set the initial values
	var typed string
	con.Active = true

	// Loop while the value is true
	for con.Active {
		// Prompt the user for a keyword
		fmt.Print(con.Prompt)
		typed = Readline()

		// If at least a character is typed
		if arr := strings.Fields(typed); len(arr) > 0 {
			// If the keyword is found
			if cmd, ok := con.commands[arr[0]]; ok {
				// Call the function
				cmd.function(typed)
				fmt.Println()
				// If the keyword is not found
			} else {
				// Output the NotFound message
				fmt.Println(con.NotFound + arr[0])
			}

			// Prevent adding an extra space after exit
			if con.Active {
				fmt.Print(con.NewLine)
			}
		}
	}
}

// *****************************************************************************
// Console Configuration
// *****************************************************************************

// The Add function registers a new console keyword, description (used in the help keyword), and function. The function must receive a string type which is the entire string of text the user typed in before pressing Enter.
func (con *Console) Add(keyword string, description string, function func(string)) {
	con.commands[keyword] = command{keyword, description, function}
}

// The Remove function unregisters a console keyword so it cannot be called.
func (con *Console) Remove(keyword string) {
	delete(con.commands, keyword)
}

// The Clear function unregisters all the console keywords so they cannot be called.
func (con *Console) Clear() {
	con.commands = make(map[string]command)
}

// *****************************************************************************
// Helpers
// *****************************************************************************

// The Readline function waits for the user to type and then press Enter. Readline returns the typed string.
func Readline() string {
	bio := bufio.NewReader(os.Stdin)
	line, _, err := bio.ReadLine()
	if err != nil {
		fmt.Println(err)
	}
	return string(line)
}

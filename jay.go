// Package main is the entry point for the Blue Jay command-line tool called
// Jay.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blue-jay/jay/env"
	"github.com/blue-jay/jay/find"
	"github.com/blue-jay/jay/generate"
	"github.com/blue-jay/jay/lib/common"
	"github.com/blue-jay/jay/migrate"
	"github.com/blue-jay/jay/replace"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("jay", "A command-line application to build faster with Blue Jay.")

	cFind          = app.Command("find", "Search for files containing matching text.")
	cFindFolder    = cFind.Arg("folder", "Folder to search").Required().String()
	cFindText      = cFind.Arg("text", "Case-sensitive text to find.").Required().String()
	cFindExtension = cFind.Arg("extension", "File name or extension to search in. Use * as a wildcard. Directory names are not valid.").Default("*.go").String()
	cFindRecursive = cFind.Arg("recursive", "True to search in subfolders. Default: true").Default("true").Bool()
	cFindFilename  = cFind.Arg("filename", "True to include file path in results if matched. Default: false").Default("false").Bool()

	cReplace          = app.Command("replace", "Search for files containing matching text and then replace it with new text.")
	cReplaceFolder    = cReplace.Arg("folder", "Folder to search").Required().String()
	cReplaceFind      = cReplace.Arg("find", "Case-sensitive text to replace.").Required().String()
	cReplaceText      = cReplace.Arg("replace", "Text to replace with.").String()
	cReplaceExtension = cReplace.Arg("extension", "File name or extension to search in. Use * as a wildcard. Directory names are not valid.").Default("*.go").String()
	cReplaceRecursive = cReplace.Arg("recursive", "True to search in subfolders. Default: true").Default("true").Bool()
	cReplaceFilename  = cReplace.Arg("filename", "True to include file path in results if matched. Default: false").Default("false").Bool()
	cReplaceCommit    = cReplace.Arg("commit", "True to makes the changes instead of just displaying them. Default: true").Default("true").Bool()

	cEnv          = app.Command("env", "Manage the environment config file.")
	cEnvMake      = cEnv.Command("make", "Create a new env.json file.")
	cEnvKeyshow   = cEnv.Command("keyshow", "Show a new set of session keys.")
	cEnvKeyUpdate = cEnv.Command("keyupdate", "Update env.json with a new set of session keys.")

	cMigrate         = app.Command("migrate", "Migrate the database to different states using 'up' and 'down' files.")
	cMigrateMake     = cMigrate.Command("make", "Create a migration file.")
	cMigrateMakeDesc = cMigrateMake.Arg("description", "Description for the migration file. Spaces will be converted to underscores and all characters will be make lowercase.").Required().String()
	cMigrateAll      = cMigrate.Command("all", "Run all 'up' files to advance the database to the latest.")
	cMigrateReset    = cMigrate.Command("reset", "Run all 'down' files to rollback the database to empty.")
	cMigrateRefresh  = cMigrate.Command("refresh", "Run all 'down' files and then 'up' files so the database is fresh and updated.")
	cMigrateStatus   = cMigrate.Command("status", "View the last 'up' file performed on the database.")
	cMigrateUp       = cMigrate.Command("up", "Apply only the next 'up' file to the database to advance the database one iteration.")
	cMigrateDown     = cMigrate.Command("down", "Apply only the current 'down' file to the database to rollback the database one iteration.")

	cGenerate     = app.Command("generate", "Generate files from template pairs.")
	cGenerateTmpl = cGenerate.Arg("folder/template", "Template pair name. Don't include an extension.").Required().String()
	cGenerateVars = stringList(cGenerate.Arg("key:value", "Key and value required for the template pair."))
)

func main() {
	app.Version("0.5-bravo")
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')

	argList := os.Args[1:]
	arg := kingpin.MustParse(app.Parse(argList))

	commandFind(arg)
	commandReplace(arg)
	commandEnv(arg)
	commandMigrate(arg, argList)
	commandGenerate(arg, argList)
}

func commandFind(arg string) {
	switch arg {
	case cFind.FullCommand():
		err := find.Run(cFindText,
			cFindFolder,
			cFindExtension,
			cFindRecursive,
			cFindFilename)
		if err != nil {
			app.Fatalf("%v", err)
		}
	}
}

func commandReplace(arg string) {
	switch arg {
	case cReplace.FullCommand():
		err := replace.Run(cReplaceFind,
			cReplaceFolder,
			cReplaceText,
			cReplaceExtension,
			cReplaceRecursive,
			cReplaceFilename,
			cReplaceCommit)
		if err != nil {
			app.Fatalf("%v", err)
		}
	}
}

func commandEnv(arg string) {
	switch arg {
	case cEnvMake.FullCommand():
		err := common.CopyFile("env.json.example", "env.json")
		if err != nil {
			app.Fatalf("%v", err)
		}
		err = env.UpdateFileKeys("env.json")
		if err != nil {
			app.Fatalf("%v", err)
		}

		p, err := filepath.Abs(".")
		if err != nil {
			app.Fatalf("%v", err)
		}
		config := filepath.Join(p, "env.json")
		if !common.Exists(config) {
			app.Fatalf("%v", err)
		}

		fmt.Println("File, env.json, created successfully with new session keys.")
		fmt.Println("Set your environment variable, JAYCONFIG, to:")
		fmt.Println(config)
	case cEnvKeyshow.FullCommand():
		fmt.Println("Paste these into your env.json file:")
		fmt.Printf(`    "AuthKey":"%v",`+"\n", env.EncodedKey(64))
		fmt.Printf(`    "EncryptKey":"%v",`+"\n", env.EncodedKey(32))
		fmt.Printf(`    "CSRFKey":"%v",`+"\n", env.EncodedKey(32))
	case cEnvKeyUpdate.FullCommand():
		err := env.UpdateFileKeys("env.json")
		if err != nil {
			app.Fatalf("%v", err)
		}
		fmt.Println("Session keys updated in env.json.")
	}
}

func commandMigrate(arg string, argList []string) {
	if argList[0] != "migrate" {
		return
	}

	mig, err := migrate.Connect()
	if err != nil {
		app.Fatalf("%v", err)
	}

	switch arg {
	case cMigrateMake.FullCommand():
		err = mig.Create(*cMigrateMakeDesc)
	case cMigrateAll.FullCommand():
		err = mig.UpAll()
	case cMigrateReset.FullCommand():
		err = mig.DownAll()
	case cMigrateRefresh.FullCommand():
		if mig.Position == -1 {
			err = mig.UpAll()
		} else {
			err = mig.DownAll()
			err = mig.UpAll()
		}
	case cMigrateStatus.FullCommand():
		fmt.Println("Last migration:", mig.Status())
	case cMigrateUp.FullCommand():
		err = mig.UpOne()
	case cMigrateDown.FullCommand():
		err = mig.DownOne()
	}

	if err != nil {
		app.Fatalf("%v", err)
	} else {
		fmt.Print(mig.Output())
	}
}

func commandGenerate(arg string, args []string) {
	if args[0] != "generate" {
		return
	}

	rootFolder, _ := common.ProjectFolder()

	err := generate.Run(args[1:], rootFolder)
	if err != nil {
		app.Fatalf("%v", err)
	}
}

// *****************************************************************************
// Custom Arguments
// *****************************************************************************

// StringList is a string array.
type StringList []string

// Set appends the string to the list.
func (i *StringList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// String returns the list.
func (i *StringList) String() string {
	return strings.Join(*i, " ")
}

// IsCumulative allows more than one value to be passed.
func (i *StringList) IsCumulative() bool {
	return true
}

// stringList accepts one or more strings as arguments.
func stringList(s kingpin.Settings) (target *StringList) {
	target = new(StringList)
	s.SetValue((*StringList)(target))
	return
}

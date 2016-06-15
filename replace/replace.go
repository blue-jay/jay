package replace

// Source: https://gist.github.com/tdegrunt/045f6b3377f3f7ffa408

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blue-jay/jay/command"
)

var (
	flagFind      *string
	flagReplace   *string
	flagExt       *string
	flagName      *bool
	flagRecursive *bool
	flagCommit    *bool
)

var Cmd = &command.Info{
	Run:       runReplace,
	UsageLine: "replace -find text -replace text -extension text [-filename=bool] [recursive=bool] [write=bool]",
	Short:     "replace text in a file",
	Long: `
Replace will find all case-sensitive strings matching -find and will replace them
with the -replace string.

Examples:
	jay replace red blue
		Replace the word "red" with the word "blue" in all go files and in child directories.
		Don't change filenames.		
	jay replace -find red -replace blue -extension "*.go" -filename=true -recursive=true -write=true
		Replace the word "red" with the word "blue" in *.go files including filenames and in child directories.
	jay replace -find "blue-jay/blueprint" -replace "user/project" -extension="*.go" -filename=true -recursive=true -write=true
		Change the name of the project and all files so it will work in another repository.
Flags:
	-find 'text to find'
		Case-sensitive text to find.
	-replace 'text to replace'
		Case-sensitive text to replace.
	-extension 'file name or just extension'
		File name or extension to modify. Use * as a wildcard. Directory names are not valid.
	-filename=bool
		True to change file names as well as text inside.
	-recursive=bool
		True to search in child directories.
	-write=bool
		True to makes changes to files instead of just analysing.
`,
}

func runReplace(cmd *command.Info, args []string) {
	flagFind = cmd.Flag.String("find", "", "search for text")
	flagReplace = cmd.Flag.String("replace", "", "replace with text")
	flagExt = cmd.Flag.String("extension", "*.go", "file extension")
	flagName = cmd.Flag.Bool("filename", false, "include file path when replacing")
	flagRecursive = cmd.Flag.Bool("recursive", true, "search all child folders")
	flagCommit = cmd.Flag.Bool("write", true, "write the changes")
	cmd.Flag.Parse(args)

	// Find or Replace is allowed to be empty, but not both
	if *flagFind != "" || *flagReplace != "" {
		replaceLogic()
	} else if len(args) == 2 {
		*flagFind = args[0]
		*flagReplace = args[1]
		replaceLogic()
	} else {
		fmt.Println("Flags are missing.")
	}
}

func replaceLogic() {
	fmt.Println()
	if *flagCommit {
		fmt.Println("Replace Results")
		fmt.Println("===============")
	} else {
		fmt.Println("Replace Results (no changes)")
		fmt.Println("============================")
	}
	err := filepath.Walk(".", replaceVisit)
	if err != nil {
		panic(err)
	}
}

func replaceVisit(path string, fi os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// If path is a folder
	if fi.IsDir() {
		// Ignore dot
		if fi.Name() == "." {
			return nil
		}

		// If recursive is true
		if *flagRecursive {
			return nil
		}

		// Don't walk the folder
		return filepath.SkipDir
	}

	matched, err := filepath.Match(*flagExt, fi.Name())

	if err != nil {
		return err
	}

	// If the file extension matches
	if matched {
		// Read the entire file into memory
		read, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println("**ERROR: Could not read from", path)
			return nil
		}

		// Convert the bytes array into a string
		oldContents := string(read)

		// If the file name contains the search term, replace the file name
		if *flagName && strings.Contains(fi.Name(), *flagFind) {
			//TODO Fix the bug where if the folder AND file name match, it won't be changed
			// Only change the filename, not the folder, or rename?
			oldpath := path
			path = strings.Replace(path, *flagFind, *flagReplace, -1)

			fmt.Println(" Rename:", oldpath, "("+path+")")

			if *flagCommit {
				errRename := os.Rename(oldpath, path)
				if errRename != nil {
					fmt.Println("**ERROR: Could not rename", oldpath, "to", path)
					return nil
				}
			}
		}

		// If the file contains the search term
		if strings.Contains(oldContents, *flagFind) {

			// Replace the search term
			newContents := strings.Replace(oldContents, *flagFind, *flagReplace, -1)

			count := strconv.Itoa(strings.Count(oldContents, *flagFind))

			fmt.Println("Replace:", path, "("+count+")")

			// Write the data back to the file
			if *flagCommit {
				err = ioutil.WriteFile(path, []byte(newContents), 0)
				if err != nil {
					fmt.Println("**ERROR: Could not write to", path)
					return nil
				}
			}
		}
	}

	return nil
}

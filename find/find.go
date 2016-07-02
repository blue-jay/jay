package find

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
	flagExt       *string
	flagName      *bool
	flagRecursive *bool
)

var Cmd = &command.Info{
	Run:       runFind,
	UsageLine: "find -find text -extension text [-filename=bool] [recursive=bool]",
	Short:     "find text in a file",
	Long: `
Find will find all case-sensitive strings matching -find.

Examples:
	jay find red
		Find the word "red" in all go files and in child directories.
	jay find red "*.*"
		Find the word "red" in all files and in child directories.
	jay replace -find red -extension "*.go" -filename=false -recursive=true
		Find word "red" in *.go files and in child directories, but excluding filenames.
Flags:
	-find 'text to find'
		Case-sensitive text to find.
	-extension 'file name or just extension'
		File name or extension to modify. Use * as a wildcard. Directory names are not valid.
	-filename=bool
		True to change file names as well as text inside.
	-recursive=bool
		True to search in child directories.
`,
}

func runFind(cmd *command.Info, args []string) {
	flagFind = cmd.Flag.String("find", "", "search for text")
	flagExt = cmd.Flag.String("extension", "*.go", "file extension")
	flagName = cmd.Flag.Bool("filename", true, "include file path when replacing")
	flagRecursive = cmd.Flag.Bool("recursive", true, "search all child folders")
	cmd.Flag.Parse(args)

	// Find is not allowed to be empty
	if *flagFind != "" {
		findLogic()
	} else if len(args) == 1 {
		*flagFind = args[0]
		findLogic()
	} else if len(args) == 2 {
		*flagFind = args[0]
		*flagExt = args[1]
		findLogic()
	} else {
		fmt.Println("Flags are missing.")
	}
}

func findLogic() {
	fmt.Println()
	fmt.Println("Search Results")
	fmt.Println("==============")

	err := filepath.Walk(".", findVisit)
	if err != nil {
		panic(err)
	}
}

func findVisit(path string, fi os.FileInfo, err error) error {
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
			oldpath := path
			fmt.Println("Filename:", oldpath)
		}

		// If the file contains the search term
		if strings.Contains(oldContents, *flagFind) {
			count := strconv.Itoa(strings.Count(oldContents, *flagFind))
			fmt.Println("Contents:", path, "("+count+")")

		}
	}

	return nil
}

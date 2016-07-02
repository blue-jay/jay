package template

// Source: https://gist.github.com/tdegrunt/045f6b3377f3f7ffa408

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/blue-jay/blueprint/lib/jsonconfig"
	"github.com/blue-jay/blueprint/lib/view"
	"github.com/blue-jay/jay/command"
)

var (
	flagMake *string
)

var Cmd = &command.Info{
	Run:       run,
	UsageLine: "template -make filename",
	Short:     "manage templates",
	Long: `
Template will help manage the templates.

Make will create a new template file in the template folder. The template
folder is read from the config.json config file stored in the environment
variable JAYCONFIG.
It supports paths with directories like 'auth/login'. Directories will be created
automatically. The file extension should not be passed. It will be read from the
config file. Any template that already exists will be overwritten.

Examples:
	jay template make auth/login
		Creates a new file called login.tmpl in the template/auth folder.
Flags:
	-make 'file to make'
		Name of template file to make. Don't include the extension.
`,
}

type info struct {
	View view.Info `json:"View"`
}

// ParseJSON unmarshals bytes to structs
func (c *info) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}

func run(cmd *command.Info, args []string) {
	flagMake = cmd.Flag.String("make", "", "name of template to make")
	cmd.Flag.Parse(args)

	folder, config := templateFolder()

	if len(*flagMake) > 0 {
		fullpath := filepath.Join(folder, *flagMake+"."+config.View.Extension)
		templateMake(fullpath)
		fmt.Println("Template created:", fullpath)
		return
	} else if len(args) == 2 {
		if args[0] == "make" {
			fullpath := filepath.Join(folder, args[1]+"."+config.View.Extension)
			templateMake(fullpath)
			fmt.Println("Template created:", fullpath)
			return
		}
	}

	fmt.Println("Flags are missing.")
}

// templateFolder returns the template folder from the config
func templateFolder() (string, *info) {
	jc := os.Getenv("JAYCONFIG")
	if len(jc) == 0 {
		log.Fatalln("Environment variable JAYCONFIG needs to be set to the config.json file location.")
	}

	config := &info{}

	// Load the configuration file
	err := jsonconfig.Load(jc, config)
	if err != nil {
		log.Fatalln("The configuration file cannot be parsed so this command will not work.")
	}

	// Build the path
	projectRoot := filepath.Dir(jc)
	templateFolder := filepath.Join(projectRoot, config.View.Folder)

	// Check to see if the folder exists
	if !folderExists(templateFolder) {
		log.Fatalln("The template folder cannot be found.")
	}

	return templateFolder, config
}

// folderExists will exit if the folder doesn't exist
func folderExists(dir string) bool {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func templateMake(filename string) {

	dir := filepath.Dir(filename)

	// Check to see if the folder exists
	if !folderExists(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}
	}

	err := ioutil.WriteFile(filename, []byte(templateText), os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
}

var templateText = `{{define "title"}}{{end}}
{{define "head"}}{{end}}
{{define "content"}}
{{end}}
{{define "foot"}}{{end}}`

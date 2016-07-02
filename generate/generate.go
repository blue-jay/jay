package generate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/blue-jay/jay/command"
)

var (
	LoopLimit = 100
)

var Cmd = &command.Info{
	Run:       run,
	UsageLine: "generate [folder/file] key:value...",
	Short:     "code generation",
	Long: `
Generate will parse and create files from template pairs.

A template pair is a set of template files:
  * default.json - json file
  * default.gen - any type of text file

Both files are template files and are parsed using the Go text/template
package. The 'jay generate' tool loops through the first level of key pairs
for empty strings. For every empty string, an argument is required to be
passed (whether empty or not) to the 'jay generate' command.

Let's look at generate/model/default.json:
{
	"config.type": "single",
	"config.output": "model/{{.package}}/{{.package}}.go",
	"package": "",
	"table": ""
}

Let's break down this command into pieces:
jay generate model/default package:automobile table:car

Argument: 'model/default'
Specifies generate/model/default.json and generate/model/default.gen are the
template pair.

Argument: 'package:automobile'
The key, 'package', from default.json will be filled with the value:
'automobile'

Argument: 'table:car'
The key, 'table', from default.json will be filled with the value: 'car'

The .json file is actually parsed up to 100 times (LoopLimit of 100 can be
changed at the package level) to ensure all variables like '{{.package}}' are
set to the correct value.

In the first iteration of parsing, the 'package' key is set to 'car'.
In the second iteration of parsing, the '{{.package}}' variables
are set to 'car' also since the 'package' key becomes a variable.

All first level keys (info, package, table) become variables after the first
iteration of parsing so they can be used without the file. If a variable is
misspelled and is never filled, a helpful error will be displayed.

The 'output' key under 'info' is required. It should be the relative output
file path to the project root for the generated file.

The folder structure of the templates (model, controller, etc) has no effect
on the generation, it's purely to aid with organization of the template pairs.

You must store the path to the env.json file in the environment
variable: JAYCONFIG. The file is at project root that is prepended to all
relative file paths.

Examples:
  jay generate model/default package:car table:car
	Generate a new model from variables in model/default.json and applies
	those variables to model/default.gen.
  jay generate controller/default package:car url:car model:car view:car
	Generate a new controller from variables in controller/default.json
	and applies those variables to controller/default.gen.
		
Flags:
  Argument 1 - model/default or controller/default
	Relative path without an extension to the template pair. Any combination
	of folders and files can be used.
  Argument 2,3,etc - package:car
	Key pair to set in the .json file. Required for every empty key in the
	.json file.
`,
}

func run(cmd *command.Info, args []string) {
	rootFolder, _ := command.ProjectFolder()

	if len(args) >= 2 {
		// Ensure the template pair files exist
		jsonFilePath := filepath.Join(rootFolder, "generate", args[0]+".json")
		if !command.Exists(jsonFilePath) {
			log.Fatalf("File doesn't exist: %v", jsonFilePath)
		}

		argMap := argsToMap(args)

		// Get the json file as a map - not parsed
		mapFile, err := jsonFileToMap(jsonFilePath)
		if err != nil {
			log.Fatal(err)
		}

		// Generate variable map
		variableMap := generateVariableMap(mapFile, argMap)

		// Check for config type
		configType, ok := variableMap["config.type"]
		if !ok {
			log.Fatal("Key, 'config.type', is missing from the .json file")
		}

		// Handle based on config.type
		switch configType {
		case "single":
			// Template File
			genFilePath := filepath.Join(rootFolder, "generate", args[0]+".gen")

			// Generate the template
			generateSingle(rootFolder, genFilePath, variableMap)

			return
		case "collection":
			generateCollection(rootFolder, variableMap)
			return
		default:
			log.Fatalf("Value of '%v' for key 'config.type' is not supported", configType)
		}
	}

	fmt.Println("Flags are missing.")
}

func generateCollection(folderPath string, variableMap map[string]interface{}) {
	// Check for required key
	collectionRaw, ok := variableMap["config.collection"]
	if !ok {
		log.Fatal("Key, 'config.collection', is missing from the .json file")
	}

	collection, ok := collectionRaw.([]interface{})
	if !ok {
		log.Fatal("Key, 'config.collection', is not in the correct format")
	}

	// Loop through the collections
	for i, v := range collection {
		vMap, ok := v.(map[string]interface{})
		if !ok {
			log.Fatal("Values for key, 'config.collection', are not in the correct format")
		}

		for name, varArray := range vMap {
			argMap, ok := varArray.(map[string]interface{})
			if !ok {
				log.Fatalf("Item at index '%v' for key, 'config.collection', is not in the correct format", i)
			}

			// Template File
			genFilePath := filepath.Join(folderPath, "generate", name+".gen")
			jsonFilePath := filepath.Join(folderPath, "generate", name+".json")

			// Get the json file as a map - not parsed
			mapFile, err := jsonFileToMap(jsonFilePath)
			if err != nil {
				log.Fatal(err)
			}

			// Generate variable map
			variableMap := generateVariableMap(mapFile, argMap)

			// Check for config type
			configType, ok := variableMap["config.type"]
			if !ok {
				log.Fatal("Key, 'config.type', is missing from the .json file")
			}

			// Handle based on config.type
			switch configType {
			case "single":
				generateSingle(folderPath, genFilePath, variableMap)
			case "collection":
				generateCollection(folderPath, variableMap)
			default:
				log.Fatalf("Value of '%v' for key 'config.type' is not supported", configType)
			}
		}
	}
}

func generateSingle(folderPath string, genFilePath string, variableMap map[string]interface{}) {
	// Check for required key
	if _, ok := variableMap["config.output"]; !ok {
		log.Fatal("Key, 'config.output', is missing from the .json file")
	}

	// Output file
	outputRelativeFile := fmt.Sprintf("%v", variableMap["config.output"])
	outputFile := filepath.Join(folderPath, outputRelativeFile)

	// Check if the file exists
	if command.Exists(outputFile) {
		log.Fatalf("Cannot generate because file already exists: %v", outputFile)
	}

	// Check if the folder exists
	dir := filepath.Dir(outputFile)
	if !command.Exists(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// If config.parse = false
	if val, ok := variableMap["config.parse"]; ok {
		if b, _ := strconv.ParseBool(fmt.Sprintf("%v", val)); !b {
			// Don't parse template, just copy to new file
			err := toFile(genFilePath, outputFile)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Code generated:", outputFile)
			return
		}
	}

	// Parse template and write to file
	err := fromMapToFile(genFilePath, variableMap, outputFile)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Code generated:", outputFile)
}

func cloneMap(originalMap map[string]interface{}) map[string]interface{} {
	copyMap := make(map[string]interface{})

	for k, v := range originalMap {
		copyMap[k] = v
	}

	return copyMap
}

// jsonFileToMap converts json file to an interface map
func jsonFileToMap(file string) (map[string]interface{}, error) {
	// Read the config file
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Convert json to interface
	var d map[string]interface{}
	err = json.Unmarshal(b, &d)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// fromMapToFile will output a file by parsing a template and applying
// variables from an interface map
func fromMapToFile(templateFile string, d map[string]interface{}, outputFile string) error {
	// Parse the template
	t, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	// Create the output file
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Fills template with variables and writes to file
	err = t.Execute(f, d)
	if err != nil {
		return err
	}

	return nil
}

// toFile will output a file by without parsing
func toFile(templateFile string, outputFile string) error {

	data, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outputFile, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func argsToMap(args []string) map[string]interface{} {
	// Fill a new map with variables
	argMap := make(map[string]interface{})
	for _, a := range args[1:] {
		arr := strings.Split(a, ":")

		if len(arr) < 2 {
			log.Fatalf("Arg is in wrong format: %v", a)
		}

		argMap[arr[0]] = strings.Join(arr[1:], ":")
	}

	return argMap
}

// generateVariableMap returns the relative file output path and the map of
// variables
func generateVariableMap(mapFile map[string]interface{}, argMap map[string]interface{}) map[string]interface{} {
	m := cloneMap(mapFile)

	// Loop through the map to find empty variables
	for s, v := range m {
		switch t := v.(type) {
		case string:
			if len(t) == 0 {
				if val, ok := argMap[s]; ok {
					m[s] = val
				} else {
					log.Fatalf("Variable missing: %v", s)
				}
			} else { // Else delete any values that are not empty
				delete(m, s)
			}
		default:
			delete(m, s)
		}
	}

	// Look through the map of the file and update it with the variables
	for s, _ := range mapFile {
		if passedVal, ok := m[s]; ok {
			mapFile[s] = passedVal
		}
	}

	// Look through the map of the variables and overwrite it with any left over arguments
	// This allows you to overwrite config.output
	for s, v := range argMap {
		if _, ok := m[s]; !ok {
			mapFile[s] = v
		}
	}

	// Counter to prevent infinite loops
	counter := 0

	for true {
		// Convert the mapFile to bytes
		mapFileBytes, err := json.Marshal(mapFile)
		if err != nil {
			log.Fatal(err)
		}

		// Create the buffer
		buf := new(bytes.Buffer)

		// Parse the template
		t, err := template.New("").Parse(string(mapFileBytes))
		if err != nil {
			log.Fatal(err)
		}

		// Fills template with variables
		err = t.Execute(buf, m)
		if err != nil {
			log.Fatal(err)
		}

		parsedTemplate := buf.Bytes()

		// Convert the json text back to a map
		err = json.Unmarshal(parsedTemplate, &m)
		if err != nil {
			log.Fatal(err)
		}

		// If the parsed template is completely filled, then stop the run, else
		// keep running
		if !strings.Contains(string(parsedTemplate), "<no value>") {
			break
		}

		var invalidKeys []string

		// Loop through the map to find empty variables
		for s, v := range m {
			switch t := v.(type) {
			case string:
				if strings.Contains(t, "<no value>") {
					invalidKeys = append(invalidKeys, s)
					delete(m, s)
				}
			default:
				// This if statement outputs a helpful error if in a nested
				// map
				if strings.Contains(fmt.Sprintf("%v", v), "<no value>") {
					invalidKeys = append(invalidKeys, fmt.Sprintf("%v %v", s, v))
				}
				delete(m, s)
			}
		}

		counter += 1

		if counter > LoopLimit {
			log.Fatalf("Check these keys for variable mistakes: %v", invalidKeys)
			break
		}
	}

	return m
}

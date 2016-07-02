package env

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/blue-jay/jay/command"
	"github.com/gorilla/securecookie"
)

var Cmd = &command.Info{
	Run:       run,
	UsageLine: "env generate",
	Short:     "manage the env.json file",
	Long: `
Env can generate an env.json file and create a new session keys.

Examples:
  jay env make
	Create a new env.json file from the env.json.example file.
  jay env updatekeys
	Generate a new Session.SecretKey in the current env.json.
Flags:
  -make
	Create a new env.json file from env.json.example with newly generated keys.
  -showkeys
	Show new session keys on the screen.
  -updatekeys
	Replace the session keys in env.json with new keys.
`,
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func run(cmd *command.Info, args []string) {
	//rootFolder, _ := command.ProjectFolder()
	if len(args) > 0 {
		switch args[0] {
		case "make":
			err := copyFile("env.json.example", "env.json")
			if err != nil {
				log.Fatal(err)
			}
			log.Println("File, env.json, created successfully.")
			return
		case "showkeys":
			fmt.Println("Paste these into your env.json file:")
			fmt.Printf(`    "AuthKey":"%v",`+"\n", encodedKey(64))
			fmt.Printf(`    "EncryptKey":"%v",`+"\n", encodedKey(32))
			return
		case "updatekeys":
			err := updateFile("env.json")
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Keys updated!")
			return
		}
	}

	fmt.Println("Flags are missing.")
}

func copyFile(src, dst string) error {
	if command.Exists(dst) {
		return errors.New(fmt.Sprintf("File, %v, already exists.", dst))
	}

	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	data = replaceKeys(data)

	return ioutil.WriteFile(dst, data, os.ModePerm)
}

func replaceKeys(data []byte) []byte {
	data = bytes.Replace(data,
		[]byte("<16 or 32 byte string>"),
		[]byte(encodedKey(64)), -1)

	data = bytes.Replace(data,
		[]byte("<16, 24, or 32 byte string>"),
		[]byte(encodedKey(32)), -1)

	return data
}

func updateFile(src string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}

	var newFile string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		if len(newFile) > 0 {
			newFile += "\n"
		}

		if strings.Contains(scanner.Text(), `"AuthKey"`) {
			newFile += fmt.Sprintf(`    "AuthKey":"%v",`, encodedKey(64))
		} else if strings.Contains(scanner.Text(), `"EncryptKey"`) {
			newFile += fmt.Sprintf(`    "EncryptKey":"%v",`, encodedKey(32))
		} else {
			newFile += scanner.Text()
		}
	}

	if err := scanner.Err(); err != nil {
		file.Close()
		return err
	}

	file.Close()
	return ioutil.WriteFile(src, []byte(newFile), os.ModePerm)
}

func encodedKey(i int) string {
	return base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(i))
}

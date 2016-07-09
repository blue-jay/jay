// Package env creates and updates the env.json file.
package env

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gorilla/securecookie"
)

// UpdateFileKeys updates the session keys in the env.json.
func UpdateFileKeys(src string) error {
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
			newFile += fmt.Sprintf(`    "AuthKey":"%v",`, EncodedKey(64))
		} else if strings.Contains(scanner.Text(), `"EncryptKey"`) {
			newFile += fmt.Sprintf(`    "EncryptKey":"%v",`, EncodedKey(32))
		} else if strings.Contains(scanner.Text(), `"CSRFKey"`) {
			newFile += fmt.Sprintf(`    "CSRFKey":"%v",`, EncodedKey(32))
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

// EncodedKey returns a base64 encoded securecookie random key.
func EncodedKey(i int) string {
	return base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(i))
}

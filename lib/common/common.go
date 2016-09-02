// Package common contains commonly used functions.
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/blue-jay/core/storage"
)

// Config returns the storage configuration information.
func Config() (*storage.Info, error) {
	config := &storage.Info{}

	jc := os.Getenv("JAYCONFIG")
	if len(jc) == 0 {
		return config, errors.New("Environment variable JAYCONFIG needs to be set to the env.json file location.")
	}

	// Read the config file
	jsonBytes, err := ioutil.ReadFile(jc)
	if err != nil {
		return config, err
	}

	// Parse the config
	err = config.ParseJSON(jsonBytes)

	return config, err
}

// LoadConfig returns the config.
func LoadConfig(configPath string) (*storage.Info, error) {
	config := &storage.Info{}

	// Read the config file
	jsonBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	// Parse the config
	err = config.ParseJSON(jsonBytes)

	return config, err
}

// ProjectFolder returns the project folder path and config.
func ProjectFolder() (string, map[string]interface{}) {
	jc := os.Getenv("JAYCONFIG")
	if len(jc) == 0 {
		log.Fatalln("Environment variable JAYCONFIG needs to be set to the env.json file location.")
	}

	info := make(map[string]interface{})

	// Read the config file
	jsonBytes, err := ioutil.ReadFile(jc)
	if err != nil {
		log.Fatalln("The configuration file cannot be found so this command will not work.")
	}

	// Parse the config
	err = json.Unmarshal(jsonBytes, &info)
	if err != nil {
		log.Fatalln("The configuration file cannot be parsed so this command will not work.")
	}

	return filepath.Dir(jc), info
}

// Exists will return true if the file or folder exists.
func Exists(f string) bool {
	if _, err := os.Stat(f); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// CopyFile copies a file from one location to another. It will return an
// error if the file already exists in the destination.
func CopyFile(src, dst string) error {
	if Exists(dst) {
		return fmt.Errorf("File, %v, already exists.", dst)
	}

	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, data, 0644)
}

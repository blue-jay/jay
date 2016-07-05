// Package migrate manages the database migrations.
//
// You must store the path to the env.json file in the
// environment variable: JAYCONFIG
//
// Examples:
//	jay migrate make "test"	# Create new migration
//	jay migrate all         # Advance all migrations
//	jay migrate reset       # Rollback all migrations
//	jay migrate refresh     # Rollback all migrations then advance all migrations
//	jay migrate status      # See last 'up' migration
//	jay migrate up          # Apply only the next 'up' migration
//	jay migrate down        # Apply only the current 'down' migration
//
//	jay migrate make "Create user table"
//	  Creates two new files in the database/migration folder using this format:
//	    * YYYYMMDD_HHMMSS.nnnnnn_create_user_table.up.sql
//	    * YYYYMMDD_HHMMSS.nnnnnn_create_user_table.down.sql
package migrate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/blue-jay/jay/lib/database"
	"github.com/blue-jay/jay/lib/migration"
	"github.com/blue-jay/jay/lib/migration/mysql"
)

// info contains the database connection information.
type info struct {
	Database database.Info `json:"Database"`
}

// ParseJSON unmarshals bytes to structs.
func (c *info) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}

// Connect creates a connection to the database.
func Connect() (*migration.Info, error) {
	var mig *migration.Info
	var di migration.Interface
	var err error

	// Get the database/migration folder and config
	folder, config, err := databaseFolder()
	if err != nil {
		return mig, err
	}

	// Determine which data to migrate
	// TODO Add other languages here
	switch config.Database.Type {
	case database.TypeMySQL:
		// Create MySQL entity
		my := &mysql.Entity{}

		// Update the config
		my.UpdateConfig(&config.Database)
		di = my

		// Connect to the database
		err = database.Connect(config.Database, true)
		if err != nil {
			// Close the open connection (since 'unknown database' is still an active
			// connection)
			database.Disconnect()

			// Connect to database without a database
			err = database.Connect(config.Database, false)
			if err != nil {
				return mig, err
			}

			// Create the database
			err = database.Create(config.Database.MySQL)
			if err != nil {
				return mig, err
			}

			// Close connection
			database.Disconnect()

			// Reconnect to the database
			err = database.Connect(config.Database, true)
			if err != nil {
				return mig, err
			}
		}
	default:
		return mig, fmt.Errorf("No registered database in config of type: %v", config.Database.Type)
	}

	// Setup logic was here
	return migration.New(di, folder)
}

// databaseFolder returns the database folder from the config.
func databaseFolder() (string, *info, error) {
	config := &info{}

	jc := os.Getenv("JAYCONFIG")
	if len(jc) == 0 {
		return "", config, errors.New("Environment variable JAYCONFIG needs to be set to the env.json file location.")
	}

	// Load the configuration file
	err := load(jc, config)
	if err != nil {
		return "", config, errors.New("The configuration file cannot be parsed so this command will not work.")
	}

	// Build the path
	projectRoot := filepath.Dir(jc)
	folder := filepath.Join(projectRoot, "database", "migration")

	// Check to see if the folder exists
	if !folderExists(folder) {
		return "", config, errors.New("The database/migration folder cannot be found.")
	}

	return folder, config, nil
}

// folderExists will exit if the folder doesn't exist.
func folderExists(dir string) bool {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Parser must implement ParseJSON.
type Parser interface {
	ParseJSON([]byte) error
}

// load the JSON config file.
func load(configFile string, p Parser) error {
	// Read the config file
	jsonBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	// Parse the config
	if err := p.ParseJSON(jsonBytes); err != nil {
		return err
	}

	return nil
}

package migrate

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/blue-jay/blueprint/lib/database"
	"github.com/blue-jay/blueprint/lib/jsonconfig"
	"github.com/blue-jay/blueprint/lib/migration"
	"github.com/blue-jay/blueprint/lib/migration/mysql"
	"github.com/blue-jay/jay/command"
)

var (
	ErrFlagsIncorrect = errors.New("Flags are incorrect.")

	flagMake     *string
	flagReset    *bool
	flagRefresh  *bool
	flagStatus   *bool
	flagUp       *bool
	flagDown     *bool
	templateUp   = ``
	templateDown = ``
)

var Cmd = &command.Info{
	Run:       run,
	UsageLine: "migrate [make filename] [reset] [refresh] [status] [up] [down]",
	Short:     "manage database migrations",
	Long: `
Migrate helps manage the database migrations.

You must store the path to the config.json file in the
environment variable: JAYCONFIG

Examples:
	jay migrate             # Apply all migrations
	jay migrate reset       # Rollback everything
	jay migrate refresh     # Rollback everything then run through all
	jay migrate status      # See current migration level of database
	jay migrate up          # Undo last migration (1 file at a time)
	jay migrate down        # Run next migration (1 file at a time)
	jay migrate make "test"	# Create new migration
	
	jay migrate make "Create user table"
	  Creates two new files in the database/migration folder using this format:
	    * YYYYMMDD_HHMMSS_create_user_table.up.sql
	    * YYYYMMDD_HHMMSS_create_user_table.down.sql
	
Flags:
	[none]
		Apply all database migrations
	-make 'description'
		Make migration files (up and down) with the description.
	-reset
		Rollback all database migrations.
	-refresh
		Rollback and then apply all database migrations.
	-status
		Display the last migration applied.
	-up
		Apply one database migration
	-down
		Rollback one database migration.
`,
}

func run(cmd *command.Info, args []string) {
	flagMake = cmd.Flag.String("make", "", "description of the migration")
	flagReset = cmd.Flag.Bool("reset", false, "rollback all migrations")
	flagRefresh = cmd.Flag.Bool("refresh", false, "rollback all and then run all migrations")
	flagStatus = cmd.Flag.Bool("status", false, "view last migration")
	flagUp = cmd.Flag.Bool("up", false, "apply next migration")
	flagDown = cmd.Flag.Bool("down", false, "rollback last migration")
	cmd.Flag.Parse(args)

	// Get the database/migration folder and config
	folder, config := databaseFolder()

	var di migration.Interface

	// Determine which data to migrate
	// TODO Add other languages here
	switch config.Database.Type {
	case database.TypeMySQL:
		// Create MySQL entity
		my := &mysql.Entity{}
		// Update the config
		my.UpdateConfig(&config.Database)
		di = my
	default:
		log.Fatal("No registered database in config of type:", config.Database.Type)
	}

	// Connect to database
	err := database.Connect(config.Database)
	if err != nil {
		log.Fatal(err)
	}

	// Run the logic after db is set
	migrateLogic(di, folder, args)
}

// Handles all the logic for the migrations
func migrateLogic(di migration.Interface, folder string, args []string) {
	// Setup logic was here
	mig, err := migration.New(di, folder)
	if err != nil {
		log.Fatal(err)
	}

	// If just migrate by itself
	if len(args) == 0 {
		handleError(mig.UpAll())
	} else if len(args) == 2 {
		arg := args[0]

		if len(*flagMake) > 0 {
			handleError(mig.Create(*flagMake))
		} else if arg == "make" {
			handleError(mig.Create(args[1]))
		} else {
			handleError(ErrFlagsIncorrect)
		}
	} else if len(args) == 1 {
		arg := args[0]

		if *flagReset || arg == "reset" {
			handleError(mig.DownAll())
		} else if *flagRefresh || arg == "refresh" {
			if mig.Position == -1 {
				handleError(mig.UpAll())
			} else {
				handleError(mig.DownAll())
				handleError(mig.UpAll())
			}
		} else if *flagStatus || arg == "status" {
			fmt.Println("Last migration:", mig.Status())
		} else if *flagUp || arg == "up" {
			handleError(mig.UpOne())
		} else if *flagDown || arg == "down" {
			handleError(mig.DownOne())
		} else {
			handleError(ErrFlagsIncorrect)
		}
	} else {
		handleError(ErrFlagsIncorrect)
	}

	fmt.Print(mig.Output())
}

func handleError(err error) {
	if err == migration.ErrNone || err == migration.ErrCurrent {
		fmt.Println(err)
	} else if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// databaseFolder returns the database folder from the config
func databaseFolder() (string, *info) {
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
	projectRoot := filepath.Dir(filepath.Dir(jc))
	folder := filepath.Join(projectRoot, "database", "migration")

	// Check to see if the folder exists
	if !folderExists(folder) {
		log.Fatalln("The database/migration folder cannot be found.")
	}

	return folder, config
}

// info contains the database connection information
type info struct {
	Database database.Info `json:"Database"`
}

// ParseJSON unmarshals bytes to structs
func (c *info) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
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

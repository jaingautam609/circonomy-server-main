package database

import (
	"fmt"
	"os"
	"path/filepath"

	migrator "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // import for migration file
	"github.com/jmoiron/sqlx"
	_ "github.com/mattes/migrate/source/file" // import for migration file
)

var (
	CirconomyDB *sqlx.DB
)

type SSLMode string

const (
	SSLModeDisable SSLMode = "disable"
)

// ConnectAndMigrate function connects with a given database and returns error if there is any error
func ConnectAndMigrate(host, port, databaseName, user, password string, sslMode SSLMode) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, databaseName, sslMode)
	DB, err := sqlx.Open("postgres", connStr)

	if err != nil {
		return err
	}

	err = DB.Ping()
	if err != nil {
		return err
	}
	CirconomyDB = DB
	return migrateUp(DB)
}

//func ShutdownDatabase() error {
//	return CirconomyDB.Close()
//}

// migrateUp function migrate the database and handles the migration logic
func migrateUp(db *sqlx.DB) error {
	db.Driver()
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	path := findMigrationsFolderRoot()
	m, err := migrator.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", path),
		os.Getenv("DB_NAME"), driver)

	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrator.ErrNoChange {
		return err
	}
	return nil
}

func findMigrationsFolderRoot() string {
	workingDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	lastDir := workingDirectory
	myUniqueRelativePath := "database/migrations"
	for {
		currentPath := fmt.Sprintf("%s/%s", lastDir, myUniqueRelativePath)
		fi, statErr := os.Stat(currentPath)
		if statErr == nil {
			mode := fi.Mode()
			if mode.IsDir() {
				return currentPath
			}
		}
		newDir := filepath.Dir(lastDir)
		if newDir == "/" || newDir == lastDir {
			return ""
		}
		lastDir = newDir
	}
}

package lib

import (
	"io/ioutil"
	"path"
	"sort"
	"strings"
	"database/sql"
	_ "github.com/lib/pq"
)

type Migration struct {
	Name    string
	UpSQL   string
	DownSQL string
}

func ReadMigrationsDir(dir string) ([]*Migration, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	migrations := map[string]*Migration{}

	withMigration := func(name string) *Migration {
		m, ok := migrations[name]
		if !ok {
			m = &Migration{
				Name: name,
			}
			migrations[name] = m
		}
		return m
	}

	// Load all migration files into migrations map
	for _, file := range files {
		filePath := path.Join(dir, file.Name())
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		name, isUp := parseMigrationFileName(file.Name())
		migration := withMigration(name)
		if isUp {
			migration.UpSQL = string(bytes)
		} else {
			migration.DownSQL = string(bytes)
		}
	}

	// Sort keys lexicographically
	keys := []string{}
	for k := range migrations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Make result slice
	result := []*Migration{}
	for _, k := range keys {
		result = append(result, migrations[k])
	}
	return result, nil
}

func parseMigrationFileName(fileName string) (string, bool) {
	return getMigrationName(fileName), getUpness(fileName)
}

func getMigrationName(fileName string) string {
	dotParts := strings.Split(fileName, ".")
	return dotParts[0]
}

func getUpness(fileName string) bool {
	return !strings.HasSuffix(fileName, ".down.sql")
}

func RunMigrations(dir string, connectionString string) error {
	migrations, err := ReadMigrationsDir(dir)
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	err = requireMigrationsTable()
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		err = execMigration(migration)
		if err != nil {
			return err
		}
	}

	return nil
}

func requireMigrationsTable(db *sql.DB) error {
	sql := "CREATE TABLE IS NOT EXISTS migrations (version INT PRIMARY KEY, at TIMESTAMP WITH TIME ZONE)"
	// TO DO...
}

func execMigration(db *sql.DB, migration Migration) error {
	// TO DO...
}
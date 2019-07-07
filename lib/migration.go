package lib

import (
	"context"
	"database/sql"
	"io/ioutil"
	"log"
	"path"
	"sort"
	"strings"

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

func RunMigrations(ctx context.Context, dir string, connectionString string) error {
	migrations, err := ReadMigrationsDir(dir)
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	err = requireMigrationsTable(ctx, db)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		// Avoid running the same migration twice
		alreadyRun, err := hasMigrationRun(ctx, db, migration)
		if err != nil {
			return err
		}
		if alreadyRun {
			continue
		}

		// Actually run migration
		err = execMigration(ctx, db, migration)
		if err != nil {
			return err
		}
	}

	return nil
}

func requireMigrationsTable(ctx context.Context, db *sql.DB) error {
	q := "CREATE TABLE IF NOT EXISTS migrations (key VARCHAR(200) PRIMARY KEY, at TIMESTAMP WITH TIME ZONE)"
	_, err := db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func execMigration(ctx context.Context, db *sql.DB, migration *Migration) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, migration.UpSQL)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Printf("Rollback failed on migration error. Rollback error: %v", rollbackErr)
		}
		return err
	}

	insertSQL := "INSERT INTO migrations (key, at) VALUES ($1, CURRENT_TIMESTAMP)"
	_, err = tx.ExecContext(ctx, insertSQL, migration.Name)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Printf("Rollback failed on migration error. Rollback error: %v", rollbackErr)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func hasMigrationRun(ctx context.Context, db *sql.DB, migration *Migration) (bool, error) {
	q := "SELECT 1 FROM migrations WHERE key=$1"
	row := db.QueryRowContext(ctx, q, migration.Name)
	result := 0
	err := row.Scan(&result)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

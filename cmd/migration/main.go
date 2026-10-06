package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

const migrationsDir = "migrations"

// migration is one migrations/v<version>__<name>.sql file.
type migration struct {
	version int
	name    string // file name, e.g. "v1__init.sql"
}

func main() {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dsn())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	paths, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("Failed to list migration files: %v", err)
	}
	if len(paths) == 0 {
		log.Fatalf("No migration files found in %s/", migrationsDir)
	}
	names := make([]string, len(paths))
	for i, p := range paths {
		names[i] = filepath.Base(p)
	}
	migrations, err := parseMigrations(names)
	if err != nil {
		log.Fatalf("Invalid migration files: %v", err)
	}

	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		log.Fatalf("Failed to read applied migrations: %v", err)
	}

	count := 0
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		if err := apply(ctx, conn, m); err != nil {
			log.Fatalf("Failed to run migration %s: %v", m.name, err)
		}
		log.Printf("Applied %s", m.name)
		count++
	}

	if count == 0 {
		log.Println("Nothing to apply")
	}
	log.Println("Migration done")
}

// parseMigrations turns file names into migrations sorted by version. It
// rejects names not shaped "v<number>__<name>.sql" and duplicate versions.
func parseMigrations(names []string) ([]migration, error) {
	seen := make(map[int]string)
	var out []migration

	for _, name := range names {
		versionPart, rest, ok := strings.Cut(strings.TrimPrefix(name, "v"), "__")
		if !ok || !strings.HasPrefix(name, "v") || !strings.HasSuffix(rest, ".sql") || rest == ".sql" {
			return nil, fmt.Errorf("%q is not named v<number>__<name>.sql", name)
		}
		version, err := strconv.Atoi(versionPart)
		if err != nil || version < 1 {
			return nil, fmt.Errorf("%q has an invalid version %q", name, versionPart)
		}
		if other, dup := seen[version]; dup {
			return nil, fmt.Errorf("%q and %q share version %d", other, name, version)
		}
		seen[version] = name
		out = append(out, migration{version: version, name: name})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

// appliedVersions creates schema_migrations if needed and returns the versions
// recorded in it. A database created before schema_migrations existed already
// has the v1 tables, so v1 is recorded there without being run.
func appliedVersions(ctx context.Context, conn *pgx.Conn) (map[int]bool, error) {
	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INT PRIMARY KEY,
			name       TEXT NOT NULL,
			applied_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return nil, err
	}

	ct, err := conn.Exec(ctx, `
		INSERT INTO schema_migrations (version, name)
		SELECT 1, 'v1__init.sql'
		WHERE NOT EXISTS (SELECT 1 FROM schema_migrations)
		  AND to_regclass('public.users') IS NOT NULL
	`)
	if err != nil {
		return nil, err
	}
	if ct.RowsAffected() > 0 {
		log.Println("Existing database found; recorded v1__init.sql as already applied")
	}

	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// apply runs one migration file and records it, in a single transaction.
func apply(ctx context.Context, conn *pgx.Conn, m migration) error {
	sql, err := os.ReadFile(filepath.Join(migrationsDir, m.name))
	if err != nil {
		return err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO schema_migrations (version, name) VALUES ($1, $2)
	`, m.version, m.name); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func dsn() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if host == "" || port == "" || user == "" || password == "" || name == "" {
		log.Fatal("Missing required database environment variables")
	}

	return "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + name + " sslmode=disable"
}

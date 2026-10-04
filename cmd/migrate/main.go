// Command migrate applies or rolls back the embedded SQL migrations.
//
// Usage:
//
//	migrate -direction up              apply every pending migration
//	migrate -direction up -steps 1     apply the next pending migration
//	migrate -direction down -steps 1   roll back the most recent migration
package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/database"
	"github.com/campuscare/api/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

type migration struct {
	version int64
	name    string
	upSQL   string
	downSQL string
}

func main() {
	// The direction may be given as a flag or as the single positional argument
	// (`migrate up -steps 1`). It is extracted before flag parsing because
	// flag.Parse stops at the first non-flag argument, which would otherwise
	// turn every flag after the direction into a stray positional argument.
	// valueFlags take a separate argument that must not be mistaken for the
	// direction, so they are re-joined with their value during the split.
	valueFlags := map[string]bool{"steps": true, "direction": true}

	args := os.Args[1:]
	positional := make([]string, 0, 1)
	flagArgs := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}

		flagArgs = append(flagArgs, arg)
		name := strings.TrimLeft(arg, "-")
		if strings.Contains(name, "=") || !valueFlags[name] {
			continue
		}
		if i+1 < len(args) {
			flagArgs = append(flagArgs, args[i+1])
			i++
		}
	}

	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	direction := fs.String("direction", "", "migration direction: up or down (may also be given positionally)")
	steps := fs.Int("steps", 0, "number of migrations to apply (0 means all pending)")
	if err := fs.Parse(flagArgs); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	if len(positional) > 1 {
		log.Fatalf("migrate: unexpected extra arguments: %v", positional[1:])
	}

	resolved := *direction
	if len(positional) == 1 {
		if resolved != "" && resolved != positional[0] {
			log.Fatalf("migrate: conflicting direction: -direction=%q and %q", resolved, positional[0])
		}
		resolved = positional[0]
	}

	if resolved == "" {
		resolved = "up"
	}

	if err := run(resolved, *steps); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}

func run(direction string, steps int) error {
	if direction != "up" && direction != "down" {
		return fmt.Errorf("invalid direction %q: expected \"up\" or \"down\"", direction)
	}

	all, err := loadMigrations()
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	pool, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		return fmt.Errorf("cannot connect to database: %w", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := ensureVersionTable(ctx, pool); err != nil {
		return err
	}

	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return err
	}

	if direction == "up" {
		return migrateUp(ctx, pool, all, applied, steps)
	}
	return migrateDown(ctx, pool, all, applied, steps)
}

func migrateUp(ctx context.Context, pool *pgxpool.Pool, all []migration, applied map[int64]string, steps int) error {
	pending := make([]migration, 0, len(all))
	for _, m := range all {
		if _, ok := applied[m.version]; !ok {
			pending = append(pending, m)
		}
	}

	if len(pending) == 0 {
		log.Println("migrate: database is already up to date")
		return nil
	}

	if steps > 0 && steps < len(pending) {
		pending = pending[:steps]
	}

	for _, m := range pending {
		log.Printf("migrate: applying %s", m.name)
		if err := applyMigration(ctx, pool, m.upSQL, m.version, true); err != nil {
			return fmt.Errorf("apply %s: %w", m.name, err)
		}
	}

	log.Printf("migrate: applied %d migration(s)", len(pending))
	return nil
}

func migrateDown(ctx context.Context, pool *pgxpool.Pool, all []migration, applied map[int64]string, steps int) error {
	pending := make([]migration, 0, len(applied))
	for _, m := range all {
		if _, ok := applied[m.version]; ok {
			pending = append(pending, m)
		}
	}
	if len(pending) == 0 {
		log.Println("migrate: nothing to roll back")
		return nil
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].version > pending[j].version })

	if steps > 0 && steps < len(pending) {
		pending = pending[:steps]
	}

	for _, m := range pending {
		if strings.TrimSpace(m.downSQL) == "" {
			return fmt.Errorf("cannot roll back %s: no matching .down.sql file", m.name)
		}

		log.Printf("migrate: rolling back %s", m.name)
		if err := applyMigration(ctx, pool, m.downSQL, m.version, false); err != nil {
			return fmt.Errorf("rollback %s: %w", m.name, err)
		}
	}

	log.Printf("migrate: rolled back %d migration(s)", len(pending))
	return nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, sql string, version int64, up bool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}

	if up {
		_, err = tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`,
			version, time.Now())
	} else {
		_, err = tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, version)
	}
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func ensureVersionTable(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := pool.Exec(ctx, query)
	return err
}

func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[int64]string, error) {
	rows, err := pool.Query(ctx, `SELECT version, applied_at FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]string)
	for rows.Next() {
		var version int64
		var at time.Time
		if err := rows.Scan(&version, &at); err != nil {
			return nil, err
		}
		applied[version] = at.Format(time.RFC3339)
	}
	return applied, rows.Err()
}

// loadMigrations reads the embedded SQL files and pairs each .up.sql with its
// optional .down.sql counterpart.
func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return nil, err
	}

	byVersion := make(map[int64]*migration)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") && !strings.HasSuffix(name, ".down.sql") {
			return nil, fmt.Errorf("migration %q must be named <version>_<name>.up.sql or .down.sql", entry.Name())
		}
		base := strings.TrimSuffix(strings.TrimSuffix(name, ".sql"), "up")
		base = strings.TrimSuffix(base, "down")
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 || parts[1] == "" {
			return nil, fmt.Errorf("migration %q must be named <version>_<name>.<up|down>.sql", entry.Name())
		}

		version, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("migration %q has a non-numeric version: %w", entry.Name(), err)
		}

		content, err := migrations.FS.ReadFile(entry.Name())
		if err != nil {
			return nil, err
		}

		m, ok := byVersion[version]
		if !ok {
			m = &migration{version: version, name: base}
			byVersion[version] = m
		}

		switch {
		case strings.HasSuffix(entry.Name(), ".up.sql"):
			if m.upSQL != "" {
				return nil, fmt.Errorf("migration %s has more than one .up.sql file", m.name)
			}
			m.upSQL = string(content)
		case strings.HasSuffix(entry.Name(), ".down.sql"):
			if m.downSQL != "" {
				return nil, fmt.Errorf("migration %s has more than one .down.sql file", m.name)
			}
			m.downSQL = string(content)
		}
	}

	all := make([]migration, 0, len(byVersion))
	for _, m := range byVersion {
		if strings.TrimSpace(m.upSQL) == "" {
			return nil, fmt.Errorf("migration %s has no .up.sql file", m.name)
		}
		all = append(all, *m)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].version < all[j].version })

	if len(all) == 0 {
		fmt.Fprintln(os.Stderr, "migrate: warning: no migrations found")
	}
	return all, nil
}

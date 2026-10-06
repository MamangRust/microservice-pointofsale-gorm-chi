package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/dotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/spf13/viper"
)

const (
	dialect = "postgres"
)

var (
	flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	dir   = flags.String("dir", "", "directory with migration files (default: collect per bounded context)")
)

func main() {
	flags.Usage = usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := flags.Args()
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		flags.Usage()
		return
	}

	command := args[0]

	err := dotenv.Viper()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	if *dir != "" {
		// Explicit dir (Docker image legacy path): delegate to goose directly
		// against the generic DB_* instance.
		connStr := buildConnStr("DB")
		db, err := goose.OpenDBWithDriver(dialect, connStr)
		if err != nil {
			log.Fatalf("Error opening database: %v", err)
		}
		defer db.Close()

		if err := goose.RunContext(context.Background(), command, db, *dir, args[1:]...); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		return
	}

	// Per bounded context: each context owns a separate PostgreSQL instance, so
	// its migrations apply to that instance only. Migration files live under
	// service/<svc>/database/migration/*.sql; we collect only the files belonging
	// to the context's member services, stage them to a temp dir (goose needs a
	// plain directory), and run them in timestamp order against DB_<CTX>_*.
	for _, ctx := range database.BoundedContexts {
		prefix := database.ContextPrefix[ctx]

		matches := collectContextMigrations(ctx)
		if len(matches) == 0 {
			log.Printf("Context %s (%s): no migration files, skipping", ctx, prefix)
			continue
		}
		sort.Strings(matches)

		tmp, err := os.MkdirTemp("", fmt.Sprintf("pos-migrate-%s-", ctx))
		if err != nil {
			log.Fatalf("Failed to create temp migrations dir for %s: %v", ctx, err)
		}

		for _, m := range matches {
			data, err := os.ReadFile(m)
			if err != nil {
				os.RemoveAll(tmp)
				log.Fatalf("Failed to read %s: %v", m, err)
			}
			if err := os.WriteFile(filepath.Join(tmp, filepath.Base(m)), data, 0o644); err != nil {
				os.RemoveAll(tmp)
				log.Fatalf("Failed to stage %s: %v", filepath.Base(m), err)
			}
		}

		connStr := buildConnStr(prefix)
		db, err := goose.OpenDBWithDriver(dialect, connStr)
		if err != nil {
			os.RemoveAll(tmp)
			log.Fatalf("Error opening database for context %s: %v", ctx, err)
		}

		log.Printf("Migrating context %s (%s) — %d files in %s", ctx, prefix, len(matches), tmp)
		runErr := goose.RunContext(context.Background(), command, db, tmp, args[1:]...)

		db.Close()
		os.RemoveAll(tmp)

		if runErr != nil {
			log.Fatalf("Migration failed for context %s: %v", ctx, runErr)
		}
	}
}

// collectContextMigrations returns every migration file belonging to the
// services that own the given bounded context.
func collectContextMigrations(ctx string) []string {
	var matches []string
	for svc, svcCtx := range database.ServiceContext {
		if svcCtx != ctx {
			continue
		}
		files, err := filepath.Glob(fmt.Sprintf("service/%s/database/migration/*.sql", svc))
		if err != nil {
			log.Fatalf("Failed to glob migration files for %s: %v", svc, err)
		}
		matches = append(matches, files...)
	}
	return matches
}

// buildConnStr reads the instance connection settings for the given env prefix
// (e.g. "DB_SALES"), falling back to the generic DB_* keys when a per-context
// key is absent.
func buildConnStr(prefix string) string {
	get := func(key string) string {
		if v := viper.GetString(fmt.Sprintf("%s_%s", prefix, key)); v != "" {
			return v
		}
		return viper.GetString("DB_" + key)
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		get("HOST"), get("PORT"), get("USERNAME"), get("NAME"), get("PASSWORD"),
	)
}

func usage() {
	fmt.Println(usagePrefix)
	flags.PrintDefaults()
	fmt.Println(usageCommands)
}

var (
	usagePrefix = `Usage: migrate COMMAND
Examples:
    migrate status
`
	usageCommands = `
Commands:
    up                   Migrate the DB to the most recent version available
    up-by-one            Migrate the DB up by 1
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    reset                Roll back all migrations
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with the current timestamp
    fix                  Apply sequential ordering to migrations`
)

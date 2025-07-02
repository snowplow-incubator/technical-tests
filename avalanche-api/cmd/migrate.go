package cmd

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/snowplow-incubator/avalanche-api/pkg/config"
	"github.com/snowplow-incubator/avalanche-api/pkg/migrations"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run database migrations using goose with embedded migration files.`,
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations("up")
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the latest migration",
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations("down")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations("status")
	},
}

func runMigrations(command string) {
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	if cfg.DatabaseURI == "" {
		log.Fatal("Database URI is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}
	defer pool.Close()

	db := stdlib.OpenDBFromPool(pool)
	//nolint:errcheck
	defer db.Close()

	migrationFS, err := migrations.GetMigrations()
	if err != nil {
		log.Fatalf("Failed to get migrations: %v", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrationFS)
	if err != nil {
		log.Fatalf("Failed to create goose provider: %v", err)
	}

	switch command {
	case "up":
		_, err = provider.Up(ctx)
	case "down":
		_, err = provider.Down(ctx)
	case "status":
		status, statusErr := provider.Status(ctx)
		if statusErr != nil {
			log.Fatalf("Failed to get migration status: %v", statusErr)
		}
		for _, s := range status {
			log.Printf("Migration %s: %s", s.Source.Path, s.State)
		}
		return
	default:
		log.Fatalf("Unknown migration command: %s", command)
	}

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Printf("Migration %s completed successfully", command)
}

func RunMigrationsUp(cfg *config.Config) error {
	if cfg.DatabaseURI == "" {
		return nil
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURI)
	if err != nil {
		return err
	}
	defer pool.Close()

	db := stdlib.OpenDBFromPool(pool)
	//nolint:errcheck
	defer db.Close()

	migrationFS, err := migrations.GetMigrations()
	if err != nil {
		return err
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrationFS)
	if err != nil {
		return err
	}

	_, err = provider.Up(ctx)
	return err
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(upCmd)
	migrateCmd.AddCommand(downCmd)
	migrateCmd.AddCommand(statusCmd)
}

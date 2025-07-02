package cmd

import (
	"log"

	"github.com/snowplow-incubator/avalanche-api/pkg/api"
	"github.com/snowplow-incubator/avalanche-api/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the Avalanche API server",
	Long:  `Start the Avalanche API server with health and events endpoints.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("port", config.DefaultPort)
		viper.SetDefault("max_workers", config.DefaultMaxWorkers)
		viper.SetDefault("max_request_size", config.DefaultMaxRequestSize)
		viper.SetDefault("batch_size", config.DefaultBatchSize)
		viper.SetDefault("database_pool_size", config.DefaultDatabasePoolSize)

		var cfg config.Config
		if err := viper.Unmarshal(&cfg); err != nil {
			log.Fatalf("Failed to unmarshal config: %v", err)
		}

		if err := RunMigrationsUp(&cfg); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		server := api.NewServer(&cfg)
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

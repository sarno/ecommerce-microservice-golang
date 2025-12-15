package cmd

import (
	"log"
	"notification-service/config"
	"notification-service/internal/core/domain/models"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "db:migrate",
	Short: "Run database migrations.",
	Long:  `This command connects to the database and runs GORM's AutoMigrate to ensure the schema is up to date.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting database migration...")

		// 1. Load application configuration
		cfg := config.NewConfig()

		// 2. Establish database connection
		postgres, err := cfg.ConnectionPostgres()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		db := postgres.DB

		// 3. Run the migration
		log.Println("Migrating notification model...")
		err = db.AutoMigrate(&models.Notification{})
		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		log.Println("Database migration completed successfully!")
	},
}

func init() {
	// Add the migrate command to the root command
	rootCmd.AddCommand(migrateCmd)
}

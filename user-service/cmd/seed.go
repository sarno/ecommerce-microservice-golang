package cmd

import (
	"log"
	"user-service/config"
	"user-service/database/seeds"

	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "db:seed",
	Short: "Seed the database with initial data.",
	Long:  `This command connects to the database and runs all the required seeders to populate the tables with initial data.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting database seeding...")

		// 1. Load application configuration
		cfg := config.NewConfig()

		// 2. Establish database connection
		postgres, err := cfg.ConnectionPostgres()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		db := postgres.DB

		// 3. Run the seeders
		log.Println("Seeding users...")
		seeds.SeedUser(db)

		log.Println("Seeding roles...")
		seeds.SeedRole(db)

		log.Println("Seeding user_roles...")
		seeds.SeedUserRole(db)

		log.Println("Database seeding completed successfully!")
	},
}

func init() {
	// Add the seed command to the root command
	rootCmd.AddCommand(seedCmd)
}

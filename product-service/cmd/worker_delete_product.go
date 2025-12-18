package cmd

import (
	"fmt"
	"product-service/internal/adapter/message"

	"github.com/spf13/cobra"
)

var workerDeleteCmd = &cobra.Command{
	Use:   "worker:delete-product",
	Short: "worker",
	Long:  "Menjalankan worker untuk consume RabbitMQ dan index ke Elasticsearch",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Worker untuk Elasticsearch Indexing sedang berjalan...")
		message.StartDeleteOrderConsumer()
	},
}

func init() {
	rootCmd.AddCommand(workerDeleteCmd)
}
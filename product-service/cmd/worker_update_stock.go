package cmd

import (
	"fmt"
	"product-service/internal/adapter/message"

	"github.com/spf13/cobra"
)

var workerUpdateStockCmd = &cobra.Command{
	Use:   "worker:update-stock",
	Short: "Menjalankan worker untuk consume RabbitMQ dan update stock",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Worker untuk update stock sedang berjalan...")
		message.StartUpdateStockConsumer()
	},
}

func init() {
	rootCmd.AddCommand(workerUpdateStockCmd)
}
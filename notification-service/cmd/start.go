package cmd

import (
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "notification:start",
	Short: "Start notification",
	Run: func(cmd *cobra.Command, args []string) {
		// app.RunServer()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

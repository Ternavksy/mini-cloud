package cmd

import (
	"log"
	"mini-cloud/internal/server"
	"net/http"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	Run: func(cmd *cobra.Command, args []string) {
		server.InitStorage()
		handler := server.NewRouter()
		log.Println("server started on :8080")
		http.ListenAndServe(":8080", handler)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

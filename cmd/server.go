package cmd

import (
	"context"
	"errors"
	"log"
	"mini-cloud/internal/filecollections"
	"mini-cloud/internal/server"
	"mini-cloud/internal/storage"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := storage.NewStorage()
		if err != nil {
			log.Fatal("failed to initialize storage: ", err)
		}

		collections := filecollections.NewService(store)
		handler := server.NewRouter(store, collections)

		httpServer := &http.Server{
			Addr:    ":8080",
			Handler: handler,
		}

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		log.Println("server started on :8080")
		errCh := make(chan error, 1)
		go func() {
			errCh <- httpServer.ListenAndServe()
		}()

		select {
		case <-ctx.Done():
			log.Println("server shutdown started")
		case err := <-errCh:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal("server failed: ", err)
			}
			return
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Fatal("server shutdown failed: ", err)
		}

		if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server failed during shutdown: ", err)
		}
		log.Println("server stopped")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

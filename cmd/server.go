package cmd

import (
	"context"
	"errors"
	"log"
	"mini-cloud/internal/database"
	"mini-cloud/internal/filecollections"
	"mini-cloud/internal/server"
	"mini-cloud/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := storage.NewStorage()
		if err != nil {
			log.Fatal("failed to initialize storage: ", err)
		}

		metadataPath := os.Getenv("METADATA_DB_PATH")
		if metadataPath == "" {
			metadataPath = "./metadata.db"
		}
		metadataDB, err := database.Open(context.Background(), metadataPath)
		if err != nil {
			log.Fatal("failed to initialize metadata database: ", err)
		}
		defer metadataDB.Close()

		collections := filecollections.NewSQLService(metadataDB)
		handler := server.NewRouter(store, collections)

		httpServer := &http.Server{
			Addr:    ":8080",
			Handler: handler,
		}

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		log.Println("server started on :8080")
		group, groupCtx := errgroup.WithContext(ctx)
		shutdownStarted := make(chan struct{})

		group.Go(func() error {
			err := httpServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		})

		group.Go(func() error {
			<-groupCtx.Done()
			close(shutdownStarted)

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			return httpServer.Shutdown(shutdownCtx)
		})

		<-shutdownStarted
		log.Println("server shutdown started")

		if err := group.Wait(); err != nil {
			log.Fatal("server failed: ", err)
		}
		log.Println("server stopped")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

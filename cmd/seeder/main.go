package main

import (
   "context"
   "fmt"
   "log/slog"
   "os"

   "github.com/clintrovert/cfbd-etl/seeder/internal/db"
   "github.com/clintrovert/cfbd-go/cfbd"
   "golang.org/x/sync/errgroup"
)

func main() {
   conf := db.Config{
      DSN:                      os.Getenv("DATABASE_DSN"),
      MaxOpenConnections:       20,
      MaxIdleConnections:       10,
      MaxConnectionLifetimeMin: 30,
   }

   db, err := db.NewDatabase(conf)
   if err != nil {
      slog.Error("failed to create database connection", "err", err)
      os.Exit(1)
   }

   isInitialized, err := db.IsInitialized()
   if err != nil {
      slog.Error("failed to verify initialized status", "err", err)
      os.Exit(1)
   }

   if !isInitialized {
      if err = db.Initialize(); err != nil {
         slog.Error("failed to initialize database", "err", err)
         os.Exit(1)
      }
   }

   api, err := cfbd.New(os.Getenv("CFBD_API_KEY"))
   if err != nil {
      slog.Error("failed to create API client", "err", err)
      os.Exit(1)
   }

   g, gCtx := errgroup.WithContext(context.Background())

   // Conferences
   g.Go(func() error {
      conferences, confErr := api.GetConferences(gCtx)
      if confErr != nil {
         slog.Error("failed to get conferences", "err", confErr)
         return fmt.Errorf("failed to get conferences; %w", confErr)
      }

      if confErr = db.UpsertConferences(conferences); confErr != nil {
         slog.Error("failed to upsert conferences", "err", confErr)
         return fmt.Errorf("failed to upset conferences; %w", confErr)
      }

      return nil
   })

   if err = g.Wait(); err != nil {
      slog.Error("seed failed", "err", err)
      os.Exit(1)
   }
}

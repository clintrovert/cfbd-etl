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
   g.Go(func() error { return loadConferences(gCtx, api, db) })
   g.Go(func() error { return loadTeams(gCtx, api, db) })
   g.Go(func() error { return loadVenues(gCtx, api, db) })
   g.Go(func() error { return loadCoaches(gCtx, api, db) })

   if err = g.Wait(); err != nil {
      slog.Error("loading core tables failed", "err", err)
      os.Exit(1)
   }
}

func loadConferences(
   ctx context.Context,
   api *cfbd.Client,
   db *db.Database) error {
   conferences, err := api.GetConferences(ctx)
   if err != nil {
      slog.Error("failed to get conferences", "err", err)
      return fmt.Errorf("failed to get conferences; %w", err)
   }

   if err = db.UpsertConferences(conferences); err != nil {
      slog.Error("failed to upsert conferences", "err", err)
      return fmt.Errorf("failed to upset conferences; %w", err)
   }

   slog.Info("conferences successfully inserted")
   return nil
}

func loadVenues(
   ctx context.Context,
   api *cfbd.Client,
   db *db.Database) error {
   venues, err := api.GetVenues(ctx)
   if err != nil {
      slog.Error("failed to get venues", "err", err)
      return fmt.Errorf("failed to get venues; %w", err)
   }

   if err = db.UpsertVenues(venues); err != nil {
      slog.Error("failed to upsert venues", "err", err)
      return fmt.Errorf("failed to upsert venues; %w", err)
   }

   slog.Info("venues successfully inserted")
   return nil
}

func loadTeams(
   ctx context.Context,
   api *cfbd.Client,
   db *db.Database) error {
   teams, err := api.GetTeams(ctx, cfbd.GetTeamsRequest{})
   if err != nil {
      slog.Error("failed to get teams", "err", err)
      return fmt.Errorf("failed to get teams; %w", err)
   }

   if err = db.UpsertTeams(teams); err != nil {
      slog.Error("failed to upsert teams", "err", err)
      return fmt.Errorf("failed to upsert teams; %w", err)
   }

   slog.Info("teams successfully inserted")
   return nil
}

func loadCoaches(
   ctx context.Context,
   api *cfbd.Client,
   db *db.Database) error {
   coaches, err := api.GetCoaches(ctx, cfbd.GetCoachesRequest{})
   if err != nil {
      slog.Error("failed to get coaches", "err", err)
      return fmt.Errorf("failed to get coaches; %w", err)
   }

   if err = db.UpsertCoaches(coaches); err != nil {
      slog.Error("failed to upsert coaches", "err", err)
      return fmt.Errorf("failed to upsert coaches; %w", err)
   }

   slog.Info("coaches successfully inserted")
   return nil
}

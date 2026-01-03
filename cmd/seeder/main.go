package main

import (
   "context"
   "log/slog"
   "os"

   "github.com/clintrovert/cfbd-etl/seeder/internal"
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

   seeder := internal.Seeder{API: api, DB: db}

   // The seeding processes is broken into 6 phases based on dependencies.
   // Each phase will be concurrently executed and depend on the one before it.

   // =============================== Phase 1 ===============================
   // Phase 1 consists of tables that do not have foreign key dependencies on
   // any other table.
   phase1, phase1Ctx := errgroup.WithContext(context.Background())
   seeder.SetExecutionContext(phase1Ctx)

   phase1.Go(seeder.SeedVenues)
   phase1.Go(seeder.SeedPlayTypes)
   phase1.Go(seeder.SeedStatTypes)
   phase1.Go(seeder.SeedDraftTeams)
   phase1.Go(seeder.SeedConferences)
   phase1.Go(seeder.SeedFieldGoalEP)
   phase1.Go(seeder.SeedDraftPositions)

   if phase1Err := phase1.Wait(); phase1Err != nil {
      slog.Error("phase 1 seeding tables failed", "err", phase1Err)
      os.Exit(1)
   }

   // =============================== Phase 2 ===============================
   // phase2, phase2Ctx := errgroup.WithContext(context.Background())
   // seeder.SetExecutionContext(phase2Ctx)
   //
   // // There's technically no point to set up concurrent execution for one
   // // request but adding it here in case more seeds are added for this phase
   // // in the future.
   // phase2.Go(seeder.SeedTeams)
   //
   // if phase2Err := phase2.Wait(); phase2Err != nil {
   //    slog.Error("phase 2 seeding tables failed", "err", phase2Err)
   //    os.Exit(1)
   // }

   // =============================== Phase 3 ===============================
   // =============================== Phase 4 ===============================
   // =============================== Phase 5 ===============================
   // =============================== Phase 6 ===============================
}

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

   database, err := db.NewDatabase(conf)
   if err != nil {
      slog.Error("failed to create database connection", "err", err)
      os.Exit(1)
   }

   isInitialized, err := database.IsInitialized()
   if err != nil {
      slog.Error("failed to verify initialized status", "err", err)
      os.Exit(1)
   }

   if !isInitialized {
      if err = database.Initialize(); err != nil {
         slog.Error("failed to initialize database", "err", err)
         os.Exit(1)
      }
   }

   api, err := cfbd.New(os.Getenv("CFBD_API_KEY"))
   if err != nil {
      slog.Error("failed to create API client", "err", err)
      os.Exit(1)
   }

   seeder, err := internal.NewSeeder(database, api)
   if err != nil {
      slog.Error("failed to create seeder", "err", err)
      os.Exit(1)
   }

   ctx := context.Background()

   // The seeding processes is split into multiple phases based on dependencies.
   // Each phase will be concurrently executed and depend on the one before it.

   // =============================== Phase 1 ===============================
   phase1, phase1Ctx := errgroup.WithContext(ctx)
   seeder.SetExecutionContext(phase1Ctx)

   phase1.Go(seeder.SeedVenues)         // 1 request
   phase1.Go(seeder.SeedPlayTypes)      // 1 request
   phase1.Go(seeder.SeedStatTypes)      // 1 request
   phase1.Go(seeder.SeedDraftTeams)     // 1 request
   phase1.Go(seeder.SeedConferences)    // 1 request
   phase1.Go(seeder.SeedFieldGoalEP)    // 1 request
   phase1.Go(seeder.SeedDraftPositions) // 1 request

   if phase1Err := phase1.Wait(); phase1Err != nil {
      slog.Error("phase 1 seeding tables failed", "err", phase1Err)
      os.Exit(1)
   }

   // =============================== Phase 2 ===============================
   phase2, phase2Ctx := errgroup.WithContext(ctx)
   seeder.SetExecutionContext(phase2Ctx)

   // There's technically no point to set up concurrent execution for one
   // request but adding it here in case more seeds are added for this phase
   // in the future.
   phase2.Go(seeder.SeedTeams) // 1 request

   if phase2Err := phase2.Wait(); phase2Err != nil {
      slog.Error("phase 2 seeding tables failed", "err", phase2Err)
      os.Exit(1)
   }

   // =============================== Phase 3 ===============================
   phase3, phase3Ctx := errgroup.WithContext(ctx)
   seeder.SetExecutionContext(phase3Ctx)

   phase3.Go(seeder.SeedCalendar) // ~20 requests
   phase3.Go(seeder.SeedGames)    // ~20 requests

   if phase3Err := phase3.Wait(); phase3Err != nil {
      slog.Error("phase 3 seeding tables failed", "err", phase3Err)
      os.Exit(1)
   }

   // =============================== Phase 4 ===============================
   phase4, phase4Ctx := errgroup.WithContext(ctx)
   seeder.SetExecutionContext(phase4Ctx)

   phase4.Go(seeder.SeedDrives) // ~20 requests (takes a while though)
   phase4.Go(seeder.SeedPlays)
   phase4.Go(seeder.SeedPlayStats)
   phase4.Go(seeder.SeedGameTeamStats)
   phase4.Go(seeder.SeedGamePlayerStats)
   phase4.Go(seeder.SeedWinProbability)
   phase4.Go(seeder.SeedAdvancedBoxScore)
   phase4.Go(seeder.SeedGameWeather)
   phase4.Go(seeder.SeedGameMedia)
   phase4.Go(seeder.SeedBettingLines)

   if phase4Err := phase4.Wait(); phase4Err != nil {
      slog.Error("phase 4 seeding tables failed", "err", phase4Err)
      os.Exit(1)
   }

   // =============================== Phase 5 ===============================
   phase5, phase5Ctx := errgroup.WithContext(ctx)
   seeder.SetExecutionContext(phase5Ctx)

   phase5.Go(seeder.SeedTeamRecords)
   phase5.Go(seeder.SeedTeamTalentComposite)
   phase5.Go(seeder.SeedTeamATS)
   phase5.Go(seeder.SeedTeamSPPlus)
   phase5.Go(seeder.SeedConferenceSPPlus)
   phase5.Go(seeder.SeedTeamSRSRankings)
   phase5.Go(seeder.SeedTeamEloRankings)
   phase5.Go(seeder.SeedTeamFPIRankings)
   phase5.Go(seeder.SeedWepaTeamSeason)
   phase5.Go(seeder.SeedWepaPassing)
   phase5.Go(seeder.SeedWepaRushing)
   phase5.Go(seeder.SeedWepaKicking)
   phase5.Go(seeder.SeedReturningProduction)
   phase5.Go(seeder.SeedPortalPlayers)
   phase5.Go(seeder.SeedSeasonPlayerStats)
   phase5.Go(seeder.SeedSeasonTeamStats)
   phase5.Go(seeder.SeedRankings)

   if phase5Err := phase5.Wait(); phase5Err != nil {
      slog.Error("phase 5 seeding tables failed", "err", phase5Err)
      os.Exit(1)
   }

   // =============================== Phase 6 ===============================
   phase6, phase6Ctx := errgroup.WithContext(ctx)
   seeder.SetExecutionContext(phase6Ctx)

   phase6.Go(seeder.SeedRecruits)
   phase6.Go(seeder.SeedRecruitingRankings)
   phase6.Go(seeder.SeedDraftPicks)

   if phase6Err := phase6.Wait(); phase6Err != nil {
      slog.Error("phase 6 seeding tables failed", "err", phase6Err)
      os.Exit(1)
   }
}

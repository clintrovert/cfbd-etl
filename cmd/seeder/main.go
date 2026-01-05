package main

import (
   "context"
   "log/slog"
   "os"

   "github.com/clintrovert/cfbd-etl/seeder/internal/db"
   "github.com/clintrovert/cfbd-etl/seeder/internal/seed"
   "github.com/clintrovert/cfbd-go/cfbd"
   "golang.org/x/sync/errgroup"
   "golang.org/x/time/rate"
)

func main() {
   slog.Info("Starting CFBD Database seeder...")

   database, err := db.NewDatabase(db.Config{
      DSN:                      os.Getenv("DATABASE_DSN"),
      MaxOpenConnections:       20,
      MaxIdleConnections:       10,
      MaxConnectionLifetimeMin: 30,
   })
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

   throttle := rate.NewLimiter(rate.Limit(10), 20)

   // Rate limiter: 10 requests per second with burst of 20
   seeder, err := seed.NewSeeder(database, api, throttle)
   if err != nil {
      slog.Error("failed to create seeder", "err", err)
      os.Exit(1)
   }

   // The seeding processes is split into multiple phases based on dependencies.
   // Each phase will be concurrently executed and depend on the one before it.
   // The number of API requests for each phase should be listed in the phase
   // caption above it.
   ctx := context.Background()

   // ========================== Phase 1 (7 requests) ==========================
   slog.Info("Starting Phase 1...")
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

   slog.Info("Phase 1 Complete.")

   // ========================== Phase 2 (1 request) ===========================
   slog.Info("Starting Phase 2...")
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

   slog.Info("Phase 2 Complete.")

   // ========================= Phase 3 (~40 requests) =========================
   slog.Info("Starting Phase 3...")
   phase3, phase3Ctx := errgroup.WithContext(ctx)
   seeder.SetExecutionContext(phase3Ctx)

   phase3.Go(seeder.SeedCalendar) // ~20 requests
   phase3.Go(seeder.SeedGames)    // ~20 requests

   if phase3Err := phase3.Wait(); phase3Err != nil {
      slog.Error("phase 3 seeding tables failed", "err", phase3Err)
      os.Exit(1)
   }

   slog.Info("Phase 3 Complete.")

   // ========================= Phase 4 (~206K requests) =======================
   slog.Info("Starting Phase 4...")
   phase4, phase4Ctx := errgroup.WithContext(ctx)
   seeder.SetExecutionContext(phase4Ctx)

   phase4.Go(seeder.SeedDrives)    // 20 requests
   phase4.Go(seeder.SeedPlays)     // 400 requests
   phase4.Go(seeder.SeedPlayStats) // 400 requests
   // phase4.Go(seeder.SeedGameTeamStats)   // 400 requests
   // phase4.Go(seeder.SeedGamePlayerStats) // 400 requests
   //
   // // TODO: Introduce rate limiter to mitigate request bursts
   // phase4.Go(seeder.SeedAdvancedBoxScore) // ~41,000 requests (as of 2025)
   // phase4.Go(seeder.SeedGameWeather)      // ~41,000 requests (as of 2025)
   // phase4.Go(seeder.SeedGameMedia)        // ~41,000 requests (as of 2025)
   // phase4.Go(seeder.SeedBettingLines)     // ~41,000 requests (as of 2025)
   // phase4.Go(seeder.SeedWinProbability)   // ~41,000 requests (as of 2025)

   if phase4Err := phase4.Wait(); phase4Err != nil {
      slog.Error("phase 4 seeding tables failed", "err", phase4Err)
      os.Exit(1)
   }

   slog.Info("Phase 4 Complete.")

   // =============================== Phase 5 ===============================
   slog.Info("Starting Phase 5...")
   // phase5, phase5Ctx := errgroup.WithContext(ctx)
   // seeder.SetExecutionContext(phase5Ctx)
   //
   // phase5.Go(seeder.SeedTeamRecords)
   // phase5.Go(seeder.SeedTeamTalentComposite)
   // phase5.Go(seeder.SeedTeamATS)
   // phase5.Go(seeder.SeedTeamSPPlus)
   // phase5.Go(seeder.SeedConferenceSPPlus)
   // phase5.Go(seeder.SeedTeamSRSRankings)
   // phase5.Go(seeder.SeedTeamEloRankings)
   // phase5.Go(seeder.SeedTeamFPIRankings)
   // phase5.Go(seeder.SeedWepaTeamSeason)
   // phase5.Go(seeder.SeedWepaPassing)
   // phase5.Go(seeder.SeedWepaRushing)
   // phase5.Go(seeder.SeedWepaKicking)
   // phase5.Go(seeder.SeedReturningProduction)
   // phase5.Go(seeder.SeedPortalPlayers)
   // phase5.Go(seeder.SeedSeasonPlayerStats)
   // phase5.Go(seeder.SeedSeasonTeamStats)
   // phase5.Go(seeder.SeedRankings)
   //
   // if phase5Err := phase5.Wait(); phase5Err != nil {
   //    slog.Error("phase 5 seeding tables failed", "err", phase5Err)
   //    os.Exit(1)
   // }
   //
   // slog.Info("Phase 5 Complete.")
   //
   // // =============================== Phase 6 ===============================
   // slog.Info("Starting Phase 6...")
   // phase6, phase6Ctx := errgroup.WithContext(ctx)
   // seeder.SetExecutionContext(phase6Ctx)
   //
   // phase6.Go(seeder.SeedRecruits)
   // phase6.Go(seeder.SeedRecruitingRankings)
   // phase6.Go(seeder.SeedDraftPicks)
   //
   // if phase6Err := phase6.Wait(); phase6Err != nil {
   //    slog.Error("phase 6 seeding tables failed", "err", phase6Err)
   //    os.Exit(1)
   // }
   //
   // slog.Info("Phase 6 Complete.")
   slog.Info("Seeding process complete.")
}

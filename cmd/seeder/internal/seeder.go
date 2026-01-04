package internal

import (
   "context"
   "fmt"
   "log/slog"
   "strconv"

   "github.com/clintrovert/cfbd-etl/seeder/internal/db"
   "github.com/clintrovert/cfbd-go/cfbd"
)

// var supportedYears = []int32{
//    2005, 2006, 2007, 2008, 2009, 2010, 2011, 2012, 2013, 2014, 2015, 2016,
//    2017, 2018, 2019, 2020, 2021, 2022, 2023, 2024, 2025,
// }

var supportedYears = []int32{2024, 2025}

type Seeder struct {
   db  *db.Database
   api *cfbd.Client
   ctx context.Context
}

func NewSeeder(db *db.Database, api *cfbd.Client) (*Seeder, error) {
   return &Seeder{db: db, api: api}, nil
}

// SetExecutionContext allows the seeder's context to be mutable. This is
// technically an antipattern and the context should be passed to the individual
// seed functions, but errgroup's Go() function wants an empty function
// signature and doing this makes the code in main.go a lot more concise.
//
// This function must be called prior to any of the Seed-ing functions below.
func (s *Seeder) SetExecutionContext(ctx context.Context) {
   s.ctx = ctx
}

// SeedPlayTypes todo:describe.
func (s *Seeder) SeedPlayTypes() error {
   playTypes, err := s.api.GetPlayTypes(s.ctx)
   if err != nil {
      slog.Error("failed to get play types", "err", err)
      return fmt.Errorf("failed to get play types; %w", err)
   }

   if err = s.db.InsertPlayTypes(s.ctx, playTypes); err != nil {
      slog.Error("failed to upsert play types", "err", err)
      return fmt.Errorf("failed to upsert play types; %w", err)
   }

   slog.Info("play types successfully inserted")
   return nil
}

// SeedConferences todo:describe.
func (s *Seeder) SeedConferences() error {
   conferences, err := s.api.GetConferences(s.ctx)
   if err != nil {
      slog.Error("failed to get conferences", "err", err)
      return fmt.Errorf("failed to get conferences; %w", err)
   }

   if err = s.db.InsertConferences(s.ctx, conferences); err != nil {
      slog.Error("failed to upsert conferences", "err", err)
      return fmt.Errorf("failed to upset conferences; %w", err)
   }

   slog.Info("conferences successfully inserted")
   return nil
}

// SeedVenues todo:describe.
func (s *Seeder) SeedVenues() error {
   venues, err := s.api.GetVenues(s.ctx)
   if err != nil {
      slog.Error("failed to get venues", "err", err)
      return fmt.Errorf("failed to get venues; %w", err)
   }

   if err = s.db.InsertVenues(s.ctx, venues); err != nil {
      slog.Error("failed to upsert venues", "err", err)
      return fmt.Errorf("failed to upsert venues; %w", err)
   }

   slog.Info("venues successfully inserted")
   return nil
}

// SeedStatTypes todo:describe.
func (s *Seeder) SeedStatTypes() error {
   statCats, err := s.api.GetStatCategories(s.ctx)
   if err != nil {
      slog.Error("failed to get play types", "err", err)
      return fmt.Errorf("failed to get play types; %w", err)
   }

   if err = s.db.InsertPlayStatTypes(s.ctx, statCats); err != nil {
      slog.Error("failed to upsert play types", "err", err)
      return fmt.Errorf("failed to upsert play types; %w", err)
   }

   slog.Info("play types successfully inserted")
   return nil
}

// SeedDraftTeams todo:describe.
func (s *Seeder) SeedDraftTeams() error {
   teams, err := s.api.GetDraftTeams(s.ctx)
   if err != nil {
      slog.Error("failed to get draft teams", "err", err)
      return fmt.Errorf("failed to get draft teams; %w", err)
   }

   if err = s.db.InsertDraftTeams(s.ctx, teams); err != nil {
      slog.Error("failed to upsert draft teams", "err", err)
      return fmt.Errorf("failed to upsert draft teams; %w", err)
   }

   slog.Info("draft teams successfully inserted")
   return nil
}

// SeedDraftPositions todo:describe.
func (s *Seeder) SeedDraftPositions() error {
   positions, err := s.api.GetDraftPositions(s.ctx)
   if err != nil {
      slog.Error("failed to get draft positions", "err", err)
      return fmt.Errorf("failed to get draft positions; %w", err)
   }

   if err = s.db.InsertDraftPositions(s.ctx, positions); err != nil {
      slog.Error("failed to upsert draft teams", "err", err)
      return fmt.Errorf("failed to upsert draft teams; %w", err)
   }

   slog.Info("draft positions successfully inserted")
   return nil
}

// SeedFieldGoalEP todo:describe.
func (s *Seeder) SeedFieldGoalEP() error {
   eps, err := s.api.GetFieldGoalExpectedPoints(s.ctx)
   if err != nil {
      slog.Error("failed to get field goal ep", "err", err)
      return fmt.Errorf("failed to get field goal ep; %w", err)
   }

   if err = s.db.InsertFieldGoalEP(s.ctx, eps); err != nil {
      slog.Error("failed to insert field goal ep", "err", err)
      return fmt.Errorf("failed to insert field goal ep; %w", err)
   }

   slog.Info("field goal EP successfully inserted")
   return nil
}

func (s *Seeder) SeedTeams() error {
   teams, err := s.api.GetTeams(s.ctx, cfbd.GetTeamsRequest{})
   if err != nil {
      slog.Error("failed to get teams", "err", err)
      return fmt.Errorf("failed to get teams; %w", err)
   }

   if err = s.db.InsertTeams(s.ctx, teams); err != nil {
      slog.Error("failed to insert teams", "err", err)
      return fmt.Errorf("failed to insert teams; %w", err)
   }

   slog.Info("teams successfully inserted")
   return nil
}

func (s *Seeder) SeedCalendar() error {
   var all []*cfbd.CalendarWeek
   for _, year := range supportedYears {
      weeks, err := s.api.GetCalendar(
         s.ctx, cfbd.GetCalendarRequest{Year: year},
      )
      if err != nil {
         slog.Error(
            "failed to get calendar",
            "year", int32ToString(year),
            "err", err,
         )
         return fmt.Errorf("failed to get calendar for year %d; %w", year, err)
      }

      all = append(all, weeks...)
   }

   if err := s.db.InsertCalendarWeeks(s.ctx, all); err != nil {
      slog.Error("failed to insert calendar", "err", err)
      return fmt.Errorf("failed to insert calendar; %w", err)
   }

   return nil
}

func (s *Seeder) SeedGames() error {
   var all []*cfbd.Game
   for _, year := range supportedYears {
      weeks, err := s.api.GetGames(
         s.ctx, cfbd.GetGamesRequest{Year: year},
      )
      if err != nil {
         slog.Error(
            "failed to get games",
            "year", int32ToString(year),
            "err", err,
         )
         return fmt.Errorf("failed to get games for year %d; %w", year, err)
      }

      all = append(all, weeks...)
   }

   if err := s.db.InsertGames(s.ctx, all); err != nil {
      slog.Error("failed to insert games", "err", err)
      return fmt.Errorf("failed to insert games; %w", err)
   }

   return nil
}

func (s *Seeder) SeedDrives() error {
   totalInserted := 0

   for _, year := range supportedYears {
      drives, err := s.api.GetDrives(s.ctx, cfbd.GetDrivesRequest{Year: year})
      if err != nil {
         slog.Error(
            "failed to get drives",
            "year", int32ToString(year),
            "err", err,
         )
         return fmt.Errorf("failed to get drives for year %d; %w", year, err)
      }

      if len(drives) > 0 {
         if err := s.db.InsertDrives(s.ctx, drives); err != nil {
            slog.Error("failed to insert drives", "err", err)
            return fmt.Errorf("failed to insert drives; %w", err)
         }
         totalInserted += len(drives)
         slog.Info("inserted drives for year",
            "year", int32ToString(year),
            "count", len(drives),
            "total", totalInserted,
         )
      }
   }

   slog.Info("all drives successfully inserted", "total_count", totalInserted)
   return nil
}

func (s *Seeder) SeedPlays() error {
   totalInserted := 0

   for _, year := range supportedYears {
      // GetPlays requires both a year and a week to be specified.
      // We must query GetCalendar first to get the available weeks
      // for each year.
      calendarWeeks, err := s.api.GetCalendar(
         s.ctx, cfbd.GetCalendarRequest{Year: year},
      )
      if err != nil {
         slog.Error(
            "failed to get calendar for plays",
            "year", int32ToString(year),
            "err", err,
         )
         return fmt.Errorf("failed to get calendar for year %d; %w", year, err)
      }

      for _, week := range calendarWeeks {
         plays, err := s.api.GetPlays(s.ctx, cfbd.GetPlaysRequest{
            Year:       year,
            Week:       week.GetWeek(),
            SeasonType: week.GetSeasonType(),
         })
         if err != nil {
            slog.Error(
               "failed to get plays",
               "year", int32ToString(year),
               "week", int32ToString(week.GetWeek()),
               "season_type", week.GetSeasonType(),
               "err", err,
            )
            return fmt.Errorf("failed to get plays for year %d, week %d, season_type %s; %w",
               year, week.GetWeek(), week.GetSeasonType(), err)
         }

         if len(plays) > 0 {
            if err := s.db.InsertPlays(s.ctx, plays); err != nil {
               slog.Error("failed to insert plays", "err", err)
               return fmt.Errorf("failed to insert plays; %w", err)
            }

            totalInserted += len(plays)
            slog.Info("inserted plays",
               "year", int32ToString(year),
               "week", int32ToString(week.GetWeek()),
               "season_type", week.GetSeasonType(),
               "count", len(plays),
               "total", totalInserted,
            )
         }
      }
   }

   slog.Info("plays successfully inserted", "total_count", totalInserted)
   return nil
}

func (s *Seeder) SeedPlayStats() error {
   return nil
}

func (s *Seeder) SeedGameTeamStats() error {
   return nil
}

func (s *Seeder) SeedGamePlayerStats() error {
   return nil
}

func (s *Seeder) SeedWinProbability() error {
   return nil
}

func (s *Seeder) SeedAdvancedBoxScore() error {
   return nil
}

func (s *Seeder) SeedGameWeather() error {
   return nil
}

func (s *Seeder) SeedGameMedia() error {
   return nil
}

func (s *Seeder) SeedBettingLines() error {
   return nil
}

func (s *Seeder) SeedTeamRecords() error {
   return nil
}

func (s *Seeder) SeedTeamTalentComposite() error {
   return nil
}

func (s *Seeder) SeedTeamATS() error {
   return nil
}

func (s *Seeder) SeedTeamSPPlus() error {
   return nil
}

func (s *Seeder) SeedConferenceSPPlus() error {
   return nil
}

func (s *Seeder) SeedTeamSRSRankings() error {
   return nil
}

func (s *Seeder) SeedTeamEloRankings() error {
   return nil
}

func (s *Seeder) SeedTeamFPIRankings() error {
   return nil
}

func (s *Seeder) SeedWepaTeamSeason() error {
   return nil
}

func (s *Seeder) SeedWepaPassing() error {
   return nil
}

func (s *Seeder) SeedWepaRushing() error {
   return nil
}

func (s *Seeder) SeedWepaKicking() error {
   return nil
}

func (s *Seeder) SeedReturningProduction() error {
   return nil
}

func (s *Seeder) SeedPortalPlayers() error {
   return nil
}

func (s *Seeder) SeedSeasonPlayerStats() error {
   return nil
}

func (s *Seeder) SeedSeasonTeamStats() error {
   return nil
}

func (s *Seeder) SeedRankings() error {
   return nil
}

func (s *Seeder) SeedRecruits() error {
   return nil
}

func (s *Seeder) SeedRecruitingRankings() error {
   return nil
}

func (s *Seeder) SeedDraftPicks() error {
   return nil
}

func int32ToString(val int32) string {
   return strconv.FormatInt(int64(val), 10)
}

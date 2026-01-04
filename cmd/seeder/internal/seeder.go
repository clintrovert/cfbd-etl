package internal

import (
   "context"
   "fmt"
   "log/slog"
   "strconv"

   "github.com/clintrovert/cfbd-etl/seeder/internal/db"
   "github.com/clintrovert/cfbd-go/cfbd"
)

var supportedYears = []int32{
   2005, 2006, 2007, 2008, 2009, 2010, 2011, 2012, 2013, 2014, 2015, 2016,
   2017, 2018, 2019, 2020, 2021, 2022, 2023, 2024, 2025,
}

type Seeder struct {
   DB      *db.Database
   API     *cfbd.Client
   Context context.Context
}

func (s *Seeder) SetExecutionContext(ctx context.Context) {
   s.Context = ctx
}

// SeedPlayTypes todo:describe.
func (s *Seeder) SeedPlayTypes() error {
   playTypes, err := s.API.GetPlayTypes(s.Context)
   if err != nil {
      slog.Error("failed to get play types", "err", err)
      return fmt.Errorf("failed to get play types; %w", err)
   }

   if err = s.DB.InsertPlayTypes(s.Context, playTypes); err != nil {
      slog.Error("failed to upsert play types", "err", err)
      return fmt.Errorf("failed to upsert play types; %w", err)
   }

   slog.Info("play types successfully inserted")
   return nil
}

// SeedConferences todo:describe.
func (s *Seeder) SeedConferences() error {
   conferences, err := s.API.GetConferences(s.Context)
   if err != nil {
      slog.Error("failed to get conferences", "err", err)
      return fmt.Errorf("failed to get conferences; %w", err)
   }

   if err = s.DB.InsertConferences(s.Context, conferences); err != nil {
      slog.Error("failed to upsert conferences", "err", err)
      return fmt.Errorf("failed to upset conferences; %w", err)
   }

   slog.Info("conferences successfully inserted")
   return nil
}

// SeedVenues todo:describe.
func (s *Seeder) SeedVenues() error {
   venues, err := s.API.GetVenues(s.Context)
   if err != nil {
      slog.Error("failed to get venues", "err", err)
      return fmt.Errorf("failed to get venues; %w", err)
   }

   if err = s.DB.InsertVenues(s.Context, venues); err != nil {
      slog.Error("failed to upsert venues", "err", err)
      return fmt.Errorf("failed to upsert venues; %w", err)
   }

   slog.Info("venues successfully inserted")
   return nil
}

// SeedStatTypes todo:describe.
func (s *Seeder) SeedStatTypes() error {
   statCats, err := s.API.GetStatCategories(s.Context)
   if err != nil {
      slog.Error("failed to get play types", "err", err)
      return fmt.Errorf("failed to get play types; %w", err)
   }

   if err = s.DB.InsertPlayStatTypes(s.Context, statCats); err != nil {
      slog.Error("failed to upsert play types", "err", err)
      return fmt.Errorf("failed to upsert play types; %w", err)
   }

   slog.Info("play types successfully inserted")
   return nil
}

// SeedDraftTeams todo:describe.
func (s *Seeder) SeedDraftTeams() error {
   teams, err := s.API.GetDraftTeams(s.Context)
   if err != nil {
      slog.Error("failed to get draft teams", "err", err)
      return fmt.Errorf("failed to get draft teams; %w", err)
   }

   if err = s.DB.InsertDraftTeams(s.Context, teams); err != nil {
      slog.Error("failed to upsert draft teams", "err", err)
      return fmt.Errorf("failed to upsert draft teams; %w", err)
   }

   slog.Info("draft teams successfully inserted")
   return nil
}

// SeedDraftPositions todo:describe.
func (s *Seeder) SeedDraftPositions() error {
   positions, err := s.API.GetDraftPositions(s.Context)
   if err != nil {
      slog.Error("failed to get draft positions", "err", err)
      return fmt.Errorf("failed to get draft positions; %w", err)
   }

   if err = s.DB.InsertDraftPositions(s.Context, positions); err != nil {
      slog.Error("failed to upsert draft teams", "err", err)
      return fmt.Errorf("failed to upsert draft teams; %w", err)
   }

   slog.Info("draft positions successfully inserted")
   return nil
}

// SeedFieldGoalEP todo:describe.
func (s *Seeder) SeedFieldGoalEP() error {
   eps, err := s.API.GetFieldGoalExpectedPoints(s.Context)
   if err != nil {
      slog.Error("failed to get field goal ep", "err", err)
      return fmt.Errorf("failed to get field goal ep; %w", err)
   }

   if err = s.DB.InsertFieldGoalEP(s.Context, eps); err != nil {
      slog.Error("failed to insert field goal ep", "err", err)
      return fmt.Errorf("failed to insert field goal ep; %w", err)
   }

   slog.Info("field goal EP successfully inserted")
   return nil
}

func (s *Seeder) SeedTeams() error {
   teams, err := s.API.GetTeams(s.Context, cfbd.GetTeamsRequest{})
   if err != nil {
      slog.Error("failed to get teams", "err", err)
      return fmt.Errorf("failed to get teams; %w", err)
   }

   if err = s.DB.InsertTeams(s.Context, teams); err != nil {
      slog.Error("failed to insert teams", "err", err)
      return fmt.Errorf("failed to insert teams; %w", err)
   }

   slog.Info("teams successfully inserted")
   return nil
}

func (s *Seeder) SeedCalendar() error {
   var all []*cfbd.CalendarWeek
   for _, year := range supportedYears {
      weeks, err := s.API.GetCalendar(
         s.Context, cfbd.GetCalendarRequest{Year: year},
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

   if err := s.DB.InsertCalendarWeeks(s.Context, all); err != nil {
      slog.Error("failed to insert calendar", "err", err)
      return fmt.Errorf("failed to insert calendar; %w", err)
   }

   return nil
}

func (s *Seeder) SeedGames() error {
   var all []*cfbd.Game
   for _, year := range supportedYears {
      weeks, err := s.API.GetGames(
         s.Context, cfbd.GetGamesRequest{Year: year},
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

   if err := s.DB.InsertGames(s.Context, all); err != nil {
      slog.Error("failed to insert games", "err", err)
      return fmt.Errorf("failed to insert games; %w", err)
   }

   return nil
}

func (s *Seeder) SeedDrives() error {
   return nil
}

func (s *Seeder) SeedPlays() error {
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

func int32ToString(val int32) string {
   return strconv.FormatInt(int64(val), 10)
}

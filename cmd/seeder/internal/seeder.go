package internal

import (
   "context"
   "fmt"
   "log/slog"

   "github.com/clintrovert/cfbd-etl/seeder/internal/db"
   "github.com/clintrovert/cfbd-go/cfbd"
)

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

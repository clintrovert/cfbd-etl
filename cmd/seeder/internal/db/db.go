package db

import (
   "errors"
   "fmt"
   "log/slog"
   "strings"
   "time"

   "github.com/clintrovert/cfbd-go/cfbd"
   "google.golang.org/protobuf/types/known/structpb"
   "google.golang.org/protobuf/types/known/wrapperspb"
   "gorm.io/datatypes"
   "gorm.io/driver/postgres"
   "gorm.io/gorm"
   "gorm.io/gorm/clause"
   "gorm.io/gorm/logger"
)

var ErrDsnMissing = errors.New("dsn is required")

type Config struct {
   DSN                      string
   MaxOpenConnections       int
   MaxIdleConnections       int
   MaxConnectionLifetimeMin int
}

type Database struct {
   *gorm.DB
}

func NewDatabase(conf Config) (*Database, error) {
   if strings.TrimSpace(conf.DSN) == "" {
      slog.Error("dsn not provided")
      return nil, ErrDsnMissing
   }

   gdb, err := gorm.Open(postgres.Open(conf.DSN), &gorm.Config{
      Logger: logger.Default.LogMode(logger.Info),
   })
   if err != nil {
      slog.Error("could not open connection", "err", err.Error())
      return nil, fmt.Errorf("could not open connection; %w", err)
   }

   sqlDB, err := gdb.DB()
   if err != nil {
      slog.Error("could not init database", "err", err.Error())
      return nil, fmt.Errorf("could not init database; %w", err)
   }

   sqlDB.SetMaxOpenConns(conf.MaxOpenConnections)
   sqlDB.SetMaxIdleConns(conf.MaxIdleConnections)
   sqlDB.SetConnMaxLifetime(
      time.Duration(conf.MaxConnectionLifetimeMin) * time.Minute,
   )

   return &Database{gdb}, nil
}

func (db *Database) Initialize() error {
   // Ensure schema exists
   if err := db.Exec(`CREATE SCHEMA IF NOT EXISTS cfbd;`).Error; err != nil {
      slog.Error("could not create schema", "err", err.Error())
      return fmt.Errorf("could not create schema; %w", err)
   }

   // Core
   if err := db.AutoMigrate(
      &Venue{},
      &Conference{},
      &Team{},
   ); err != nil {
      slog.Error("could not auto-migrate core tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate core tables; %w", err)
   }

   // Simple metrics / lookups
   if err := db.AutoMigrate(
      &AdjustedTeamMetrics{},
      &PlayerWeightedEPA{},
      &KickerPAAR{},
      &TeamATS{},
      &RosterPlayer{},
      &TeamTalent{},
      &PlayerStat{},
      &TeamStat{},
   ); err != nil {
      slog.Error("could not auto-migrate metrics tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate metrics tables; %w", err)
   }

   // Matchups
   if err := db.AutoMigrate(
      &Matchup{},
      &MatchupGame{},
   ); err != nil {
      slog.Error("could not auto-migrate matchup tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate matchup tables; %w", err)
   }

   // SP / SRS / Elo / FPI
   if err := db.AutoMigrate(
      &TeamSP{},
      &ConferenceSP{},
      &TeamSRS{},
      &TeamElo{},
      &TeamFPI{},
   ); err != nil {
      slog.Error("could not auto-migrate ratings tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate ratings tables; %w", err)
   }

   // Polls
   if err := db.AutoMigrate(
      &PollWeek{},
      &Poll{},
      &PollRank{},
   ); err != nil {
      slog.Error("could not auto-migrate poll tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate poll tables; %w", err)
   }

   // Plays & stats
   if err := db.AutoMigrate(
      &PlayType{},
      &PlayStatType{},
      &Play{},
      &PlayStat{},
      &PlayerSearchResult{},
      &PlayerUsage{},
   ); err != nil {
      slog.Error("could not auto-migrate play tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate play tables; %w", err)
   }

   // Returning production & transfers
   if err := db.AutoMigrate(
      &ReturningProduction{},
      &PlayerTransfer{},
   ); err != nil {
      slog.Error("could not auto-migrate returning production tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate returning production tables; %w", err)
   }

   // Predicted points & PPA added
   if err := db.AutoMigrate(
      &PredictedPointsValue{},
      &TeamSeasonPredictedPointsAdded{},
      &TeamGamePredictedPointsAdded{},
      &PlayerGamePredictedPointsAdded{},
      &PlayerSeasonPredictedPointsAdded{},
   ); err != nil {
      slog.Error("could not auto-migrate predicted points tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate predicted points tables; %w", err)
   }

   // Win probability
   if err := db.AutoMigrate(
      &PlayWinProbability{},
      &PregameWinProbability{},
      &FieldGoalEP{},
   ); err != nil {
      slog.Error("could not auto-migrate win probability tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate win probability tables; %w", err)
   }

   // Live game
   if err := db.AutoMigrate(
      &LiveGame{},
      &LiveGameTeam{},
      &LiveGameDrive{},
      &LiveGamePlay{},
   ); err != nil {
      slog.Error("could not auto-migrate live game tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate live game tables; %w", err)
   }

   // Betting
   if err := db.AutoMigrate(
      &BettingGame{},
      &GameLine{},
      &UserInfo{},
   ); err != nil {
      slog.Error("could not auto-migrate betting tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate betting tables; %w", err)
   }

   // Games (main)
   if err := db.AutoMigrate(&Game{}); err != nil {
      slog.Error("could not auto-migrate games table", "err", err.Error())
      return fmt.Errorf("could not auto-migrate games table; %w", err)
   }

   // Game team & player stats
   if err := db.AutoMigrate(
      &GameTeamStats{},
      &GameTeamStatsTeam{},
      &GameTeamStatsTeamStat{},
      &GamePlayerStats{},
   ); err != nil {
      slog.Error("could not auto-migrate game stats tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate game stats tables; %w", err)
   }

   // Media & weather
   if err := db.AutoMigrate(
      &GameMedia{},
      &GameWeather{},
   ); err != nil {
      slog.Error("could not auto-migrate media/weather tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate media/weather tables; %w", err)
   }

   // Records / calendar / scoreboard
   if err := db.AutoMigrate(
      &TeamRecords{},
      &CalendarWeek{},
      &Scoreboard{},
   ); err != nil {
      slog.Error("could not auto-migrate records/calendar/scoreboard tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate records/calendar/scoreboard tables; %w", err)
   }

   // Drives
   if err := db.AutoMigrate(&Drive{}); err != nil {
      slog.Error("could not auto-migrate drives table", "err", err.Error())
      return fmt.Errorf("could not auto-migrate drives table; %w", err)
   }

   // Draft
   if err := db.AutoMigrate(&DraftPick{}); err != nil {
      slog.Error("could not auto-migrate draft picks table", "err", err.Error())
      return fmt.Errorf("could not auto-migrate draft picks table; %w", err)
   }

   // Coaches
   if err := db.AutoMigrate(
      &Coach{},
      &CoachSeason{},
   ); err != nil {
      slog.Error("could not auto-migrate coach tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate coach tables; %w", err)
   }

   // Recruiting
   if err := db.AutoMigrate(
      &Recruit{},
      &TeamRecruitingRanking{},
      &AggregatedTeamRecruiting{},
   ); err != nil {
      slog.Error("could not auto-migrate recruiting tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate recruiting tables; %w", err)
   }

   // Game havoc
   if err := db.AutoMigrate(&GameHavocStats{}); err != nil {
      slog.Error("could not auto-migrate game havoc stats table", "err", err.Error())
      return fmt.Errorf("could not auto-migrate game havoc stats table; %w", err)
   }

   // Advanced season stats (normalized)
   if err := db.AutoMigrate(
      &AdvRateMetrics{},
      &AdvHavoc{},
      &AdvFieldPosition{},
      &AdvSeasonStatSide{},
      &AdvancedSeasonStatsNormalized{},
   ); err != nil {
      slog.Error("could not auto-migrate advanced season stats tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate advanced season stats tables; %w", err)
   }

   // Advanced game stats (normalized)
   if err := db.AutoMigrate(
      &AdvGamePlayMetrics{},
      &AdvGameDownMetrics{},
      &AdvGameStatSide{},
      &AdvancedGameStatsNormalized{},
   ); err != nil {
      slog.Error("could not auto-migrate advanced game stats tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate advanced game stats tables; %w", err)
   }

   // Advanced box score (normalized)
   if err := db.AutoMigrate(
      &AdvancedBoxScoreGameInfo{},
      &AdvancedBoxScore{},
      &StatsByQuarter{},
      &AbsTeamFieldPosition{},
      &AbsTeamScoringOpportunities{},
      &AbsTeamHavoc{},
      &AbsTeamRushingStats{},
      &AbsTeamExplosiveness{},
      &AbsTeamSuccessRates{},
      &AbsTeamPPA{},
      &PlayerStatsByQuarter{},
      &AbsPlayerPPA{},
      &PlayerGameUsageQuarters{},
      &AbsPlayerGameUsage{},
   ); err != nil {
      slog.Error("could not auto-migrate advanced box score tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate advanced box score tables; %w", err)
   }

   // Constraints GORM doesn't create reliably: CHECK(kind IN ...)
   if err := db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'abs_team_ppa_kind_check'
			) THEN
				ALTER TABLE cfbd.abs_team_ppa
				ADD CONSTRAINT abs_team_ppa_kind_check
				CHECK (kind IN ('ppa','cumulative_ppa'));
			END IF;
		END $$;
	`).Error; err != nil {
      slog.Error("could not create abs_team_ppa kind check constraint", "err", err.Error())
      return fmt.Errorf("could not create abs_team_ppa kind check constraint; %w", err)
   }

   return nil
}

// IsInitialized returns true if the DB appears initialized.
// This checks for a reliable sentinel: presence of the cfbd schema AND
// one "late" table that only exists after a full run (abs_player_game_usage),
// plus the CHECK constraint we add manually.
func (db *Database) IsInitialized() (bool, error) {
   type row struct {
      Exists bool
   }

   // 1) schema exists?
   var schema row
   if err := db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.schemata
			WHERE schema_name = 'cfbd'
		) AS exists;
	`).Scan(&schema).Error; err != nil {
      slog.Error("could not check if schema exists", "err", err.Error())
      return false, fmt.Errorf("could not check if schema exists; %w", err)
   }
   if !schema.Exists {
      return false, nil
   }

   // 2) does a "late" table exist? (created near the end of Initialize)
   // Pick a table that only appears after most migrations ran.
   var table row
   if err := db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'cfbd'
			  AND table_name = 'abs_player_game_usage'
		) AS exists;
	`).Scan(&table).Error; err != nil {
      slog.Error("could not check if initialization table exists", "err", err.Error())
      return false, fmt.Errorf("could not check if initialization table exists; %w", err)
   }
   if !table.Exists {
      return false, nil
   }

   // 3) does the CHECK constraint exist? (we add it manually at the end)
   var constraint row
   if err := db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM pg_constraint
			WHERE conname = 'abs_team_ppa_kind_check'
		) AS exists;
	`).Scan(&constraint).Error; err != nil {
      slog.Error("could not check if check constraint exists", "err", err.Error())
      return false, fmt.Errorf("could not check if check constraint exists; %w", err)
   }
   if !constraint.Exists {
      return false, nil
   }

   return true, nil
}

func (db *Database) UpsertConferences(conferences []*cfbd.Conference) error {
   if len(conferences) == 0 {
      return nil
   }

   models := make([]Conference, 0, len(conferences))

   for _, c := range conferences {
      if c == nil {
         continue
      }

      models = append(models, Conference{
         ID:             int(c.GetId()),
         Name:           c.GetName(),
         ShortName:      stringPtr(c.GetShortName()),
         Abbreviation:   stringPtr(c.GetAbbreviation()),
         Classification: stringPtr(c.GetClassification()),
      })
   }

   if err := db.
      Clauses(clause.OnConflict{
         Columns: []clause.Column{
            {Name: "id"},
         },
         DoUpdates: clause.AssignmentColumns([]string{
            "name",
            "short_name",
            "abbreviation",
            "classification",
         }),
      }).
      CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not upsert conferences", "err", err.Error())
      return fmt.Errorf("could not upsert conferences; %w", err)
   }

   return nil
}

// UpsertTeams upserts Team rows into cfbd.teams.
// It also upserts the nested Venue (Team.location) into cfbd.venues when present,
// and sets team.venue_id accordingly.
func (db *Database) UpsertTeams(teams []*cfbd.Team) error {
   if len(teams) == 0 {
      return nil
   }

   // 1) Upsert venues first (dedupe by venue id)
   venuesByID := make(map[int]Venue)

   for _, t := range teams {
      if t == nil || t.GetLocation() == nil {
         continue
      }
      v := t.GetLocation()
      if v.GetId() == nil {
         continue // no id => cannot upsert into venues PK
      }

      id := int(v.GetId().GetValue())
      venuesByID[id] = Venue{
         ID:          id,
         Name:        stringPtr(v.GetName()),
         City:        stringPtr(v.GetCity()),
         State:       stringPtr(v.GetState()),
         Zip:         stringPtr(v.GetZip()),
         CountryCode: stringPtr(v.GetCountryCode()),
         Timezone:    stringPtr(v.GetTimezone()),
         Latitude: func() *float64 {
            if v.GetLatitude() == nil {
               return nil
            }
            x := v.GetLatitude().GetValue()
            return &x
         }(),
         Longitude: func() *float64 {
            if v.GetLongitude() == nil {
               return nil
            }
            x := v.GetLongitude().GetValue()
            return &x
         }(),
         Elevation:        stringPtr(v.GetElevation()),
         Capacity:         intPtr(v.GetCapacity()),
         ConstructionYear: intPtr(v.GetConstructionYear()),
         Grass: func() *bool {
            if v.GetGrass() == nil {
               return nil
            }
            x := v.GetGrass().GetValue()
            return &x
         }(),
         Dome: func() *bool {
            if v.GetDome() == nil {
               return nil
            }
            x := v.GetDome().GetValue()
            return &x
         }(),
      }
   }

   if len(venuesByID) > 0 {
      venueModels := make([]Venue, 0, len(venuesByID))
      for _, v := range venuesByID {
         venueModels = append(venueModels, v)
      }

      if err := db.Clauses(clause.OnConflict{
         Columns: []clause.Column{{Name: "id"}},
         DoUpdates: clause.AssignmentColumns([]string{
            "name",
            "city",
            "state",
            "zip",
            "country_code",
            "timezone",
            "latitude",
            "longitude",
            "elevation",
            "capacity",
            "construction_year",
            "grass",
            "dome",
         }),
      }).CreateInBatches(venueModels, 500).Error; err != nil {
         slog.Error("could not upsert venues", "err", err.Error())
         return fmt.Errorf("could not upsert venues; %w", err)
      }
   }

   // 2) Build team models
   teamModels := make([]Team, 0, len(teams))

   for _, t := range teams {
      if t == nil {
         continue
      }

      altNames, err := listValueJSON(t.GetAlternateNames())
      if err != nil {
         slog.Error("could not marshal team alternate_names", "err", err.Error())
         return fmt.Errorf("could not marshal team alternate_names; %w", err)
      }

      logos, err := listValueJSON(t.GetLogos())
      if err != nil {
         slog.Error("could not marshal team logos", "err", err.Error())
         return fmt.Errorf("could not marshal team logos; %w", err)
      }

      var venueID *int
      if t.GetLocation() != nil && t.GetLocation().GetId() != nil {
         id := int(t.GetLocation().GetId().GetValue())
         venueID = &id
      }

      teamModels = append(teamModels, Team{
         ID:             int(t.GetId()),
         School:         t.GetSchool(),
         Mascot:         stringPtr(t.GetMascot()),
         Abbreviation:   stringPtr(t.GetAbbreviation()),
         AlternateNames: altNames,
         Conference:     stringPtr(t.GetConference()),
         Division:       stringPtr(t.GetDivision()),
         Classification: stringPtr(t.GetClassification()),
         Color:          stringPtr(t.GetColor()),
         AlternateColor: stringPtr(t.GetAlternateColor()),
         Logos:          logos,
         Twitter:        t.GetTwitter(),
         VenueID:        venueID,
      })
   }

   // 3) Upsert teams
   if err := db.Clauses(clause.OnConflict{
      Columns: []clause.Column{{Name: "id"}},
      DoUpdates: clause.AssignmentColumns([]string{
         "school",
         "mascot",
         "abbreviation",
         "alternate_names",
         "conference",
         "division",
         "classification",
         "color",
         "alternate_color",
         "logos",
         "twitter",
         "venue_id",
      }),
   }).CreateInBatches(teamModels, 500).Error; err != nil {
      slog.Error("could not upsert teams", "err", err.Error())
      return fmt.Errorf("could not upsert teams; %w", err)
   }

   return nil
}

func (db *Database) UpsertVenues(venues []*cfbd.Venue) error {
   if len(venues) == 0 {
      return nil
   }

   models := make([]Venue, 0, len(venues))

   for _, v := range venues {
      if v == nil || v.GetId() == nil {
         // cannot upsert without primary key
         continue
      }

      models = append(models, Venue{
         ID:               int(v.GetId().GetValue()),
         Name:             stringPtr(v.GetName()),
         City:             stringPtr(v.GetCity()),
         State:            stringPtr(v.GetState()),
         Zip:              stringPtr(v.GetZip()),
         CountryCode:      stringPtr(v.GetCountryCode()),
         Timezone:         stringPtr(v.GetTimezone()),
         Latitude:         float64Ptr(v.GetLatitude()),
         Longitude:        float64Ptr(v.GetLongitude()),
         Elevation:        stringPtr(v.GetElevation()),
         Capacity:         intPtr(v.GetCapacity()),
         ConstructionYear: intPtr(v.GetConstructionYear()),
         Grass:            boolPtr(v.GetGrass()),
         Dome:             boolPtr(v.GetDome()),
      })
   }

   if len(models) == 0 {
      return nil
   }

   if err := db.Clauses(clause.OnConflict{
      Columns: []clause.Column{{Name: "id"}},
      DoUpdates: clause.AssignmentColumns([]string{
         "name",
         "city",
         "state",
         "zip",
         "country_code",
         "timezone",
         "latitude",
         "longitude",
         "elevation",
         "capacity",
         "construction_year",
         "grass",
         "dome",
      }),
   }).CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not upsert venues", "err", err.Error())
      return fmt.Errorf("could not upsert venues; %w", err)
   }

   return nil
}

func (db *Database) UpsertCoaches(coaches []*cfbd.Coach) error {
   if len(coaches) == 0 {
      return nil
   }

   // Transaction keeps coaches + seasons consistent.
   if err := db.Transaction(func(tx *gorm.DB) error {
      for _, c := range coaches {
         if c == nil {
            continue
         }
         if c.GetHireDate() == nil {
            // hire_date is NOT NULL in your schema; cannot insert without it
            continue
         }

         hireDate := c.GetHireDate().AsTime()

         // 1) Find-or-create Coach using (first_name, last_name, hire_date)
         var coach Coach
         err := tx.Where(
            "first_name = ? AND last_name = ? AND hire_date = ?",
            c.GetFirstName(),
            c.GetLastName(),
            hireDate,
         ).First(&coach).Error

         if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
               coach = Coach{
                  FirstName: c.GetFirstName(),
                  LastName:  c.GetLastName(),
                  HireDate:  hireDate,
               }
               if err := tx.Create(&coach).Error; err != nil {
                  slog.Error("could not create coach", "err", err.Error())
                  return fmt.Errorf("could not create coach; %w", err)
               }
            } else {
               slog.Error("could not query coach", "err", err.Error())
               return fmt.Errorf("could not query coach; %w", err)
            }
         }

         // 2) Upsert coach seasons (unique constraint: (coach_id, school, year))
         if len(c.GetSeasons()) == 0 {
            continue
         }

         seasons := make([]CoachSeason, 0, len(c.GetSeasons()))
         for _, s := range c.GetSeasons() {
            if s == nil {
               continue
            }
            seasons = append(seasons, CoachSeason{
               CoachID: coach.CoachID,

               School: s.GetSchool(),
               Year:   int(s.GetYear()),

               Games:  int(s.GetGames()),
               Wins:   int(s.GetWins()),
               Losses: int(s.GetLosses()),
               Ties:   int(s.GetTies()),

               PreseasonRank:  intPtr(s.GetPreseasonRank()),
               PostseasonRank: intPtr(s.GetPostseasonRank()),

               SRS:       float64Ptr(s.GetSrs()),
               SPOverall: float64Ptr(s.GetSpOverall()),
               SPOffense: float64Ptr(s.GetSpOffense()),
               SPDefense: float64Ptr(s.GetSpDefense()),
            })
         }

         if len(seasons) == 0 {
            continue
         }

         if err := tx.Clauses(clause.OnConflict{
            Columns: []clause.Column{
               {Name: "coach_id"},
               {Name: "school"},
               {Name: "year"},
            },
            DoUpdates: clause.AssignmentColumns([]string{
               "games",
               "wins",
               "losses",
               "ties",
               "preseason_rank",
               "postseason_rank",
               "srs",
               "sp_overall",
               "sp_offense",
               "sp_defense",
            }),
         }).CreateInBatches(seasons, 500).Error; err != nil {
            slog.Error("could not upsert coach seasons", "err", err.Error())
            return fmt.Errorf("could not upsert coach seasons; %w", err)
         }
      }

      return nil
   }); err != nil {
      slog.Error("could not upsert coaches", "err", err.Error())
      return fmt.Errorf("could not upsert coaches; %w", err)
   }

   return nil
}

func stringPtr(v *wrapperspb.StringValue) *string {
   if v == nil {
      return nil
   }
   s := v.GetValue()
   return &s
}

func intPtr(v *wrapperspb.Int32Value) *int {
   if v == nil {
      return nil
   }
   i := int(v.GetValue())
   return &i
}

func boolPtr(v *wrapperspb.BoolValue) *bool {
   if v == nil {
      return nil
   }
   b := v.GetValue()
   return &b
}

func float64Ptr(v *wrapperspb.DoubleValue) *float64 {
   if v == nil {
      return nil
   }
   f := v.GetValue()
   return &f
}

// ListValue -> jsonb
func listValueJSON(v *structpb.ListValue) (datatypes.JSON, error) {
   if v == nil {
      return datatypes.JSON([]byte("null")), nil
   }
   b, err := v.MarshalJSON()
   if err != nil {
      return nil, err
   }
   return datatypes.JSON(b), nil
}

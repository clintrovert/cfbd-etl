package db

import (
   "context"
   "errors"
   "fmt"
   "log/slog"
   "strings"
   "time"

   "github.com/clintrovert/cfbd-go/cfbd"
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

// Initialize creates the cfbd schema (if needed) and migrates all tables
// defined in the models/model.go I generated (package models).
//
// NOTE: Adjust the import path for your models package accordingly.
func (db *Database) Initialize() error {
   // Ensure schema exists
   if err := db.Exec(`CREATE SCHEMA IF NOT EXISTS cfbd;`).Error; err != nil {
      slog.Error("could not create schema", "err", err.Error())
      return fmt.Errorf("could not create schema; %w", err)
   }

   // ---- MIGRATION ORDER MATTERS (FKs / dependencies) ----
   // 1) Reference/dim tables first
   if err := db.AutoMigrate(
      Venue{},
      Conference{},
      Team{},
   ); err != nil {
      slog.Error("could not auto-migrate reference tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate reference tables; %w", err)
   }

   // 2) Core spine
   if err := db.AutoMigrate(
      Game{},
   ); err != nil {
      slog.Error("could not auto-migrate games table", "err", err.Error())
      return fmt.Errorf("could not auto-migrate games table; %w", err)
   }

   // 3) Matchups
   if err := db.AutoMigrate(
      Matchup{},
      MatchupGame{},
   ); err != nil {
      slog.Error("could not auto-migrate matchup tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate matchup tables; %w", err)
   }

   // 4) Calendar / scoreboard / records
   if err := db.AutoMigrate(
      CalendarWeek{},
      Scoreboard{},
      TeamRecords{},
   ); err != nil {
      slog.Error("could not auto-migrate calendar/scoreboard/records tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate calendar/scoreboard/records tables; %w", err)
   }

   // 5) Plays / drives + lookup tables
   if err := db.AutoMigrate(
      PlayType{},
      PlayStatType{},
      Drive{},
      Play{},
      PlayStat{},
   ); err != nil {
      slog.Error("could not auto-migrate play/drive tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate play/drive tables; %w", err)
   }

   // 6) Game box score stats (nested)
   if err := db.AutoMigrate(
      GameTeamStats{},
      GameTeamStatsTeam{},
      GameTeamStatsTeamStat{},

      GamePlayerStats{},
      GamePlayerStatsTeam{},
      GamePlayerStatCategories{},
      GamePlayerStatTypes{},
      GamePlayerStatPlayer{},
   ); err != nil {
      slog.Error("could not auto-migrate game stats tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate game stats tables; %w", err)
   }

   // 7) Live game (nested)
   if err := db.AutoMigrate(
      LiveGame{},
      LiveGameTeam{},
      LiveGameDrive{},
      LiveGamePlay{},
   ); err != nil {
      slog.Error("could not auto-migrate live game tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate live game tables; %w", err)
   }

   // 8) Media & weather
   if err := db.AutoMigrate(
      GameMedia{},
      GameWeather{},
   ); err != nil {
      slog.Error("could not auto-migrate media/weather tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate media/weather tables; %w", err)
   }

   // 9) Win probability
   if err := db.AutoMigrate(
      PlayWinProbability{},
      PregameWinProbability{},
      FieldGoalEP{},
   ); err != nil {
      slog.Error("could not auto-migrate win probability tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate win probability tables; %w", err)
   }

   // 10) PPA / predicted points
   if err := db.AutoMigrate(
      PredictedPointsValue{},
      TeamSeasonPredictedPointsAdded{},
      TeamGamePredictedPointsAdded{},
      PlayerGamePredictedPointsAdded{},
      PlayerSeasonPredictedPointsAdded{},
   ); err != nil {
      slog.Error("could not auto-migrate PPA tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate PPA tables; %w", err)
   }

   // 11) Advanced box score payload table (jsonb)
   if err := db.AutoMigrate(
      AdvancedBoxScore{},
   ); err != nil {
      slog.Error("could not auto-migrate advanced box score tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate advanced box score tables; %w", err)
   }

   // 12) Players / roster / usage / transfers / search
   if err := db.AutoMigrate(
      RosterPlayer{},
      PlayerSearchResult{},
      PlayerUsageSplits{},
      PlayerUsage{},
      ReturningProduction{},
      PlayerTransfer{},
      PlayerStat{},
      TeamStat{},
   ); err != nil {
      slog.Error("could not auto-migrate player tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate player tables; %w", err)
   }

   // 13) Recruiting
   if err := db.AutoMigrate(
      RecruitHometownInfo{},
      Recruit{},
      TeamRecruitingRanking{},
      AggregatedTeamRecruiting{},
   ); err != nil {
      slog.Error("could not auto-migrate recruiting tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate recruiting tables; %w", err)
   }

   // 14) Ratings
   if err := db.AutoMigrate(
      TeamSP{},
      ConferenceSP{},
      TeamSRS{},
      TeamElo{},
      TeamFPI{},
   ); err != nil {
      slog.Error("could not auto-migrate ratings tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate ratings tables; %w", err)
   }

   // 15) Polls / rankings
   if err := db.AutoMigrate(
      PollWeek{},
      Poll{},
      PollRank{},
   ); err != nil {
      slog.Error("could not auto-migrate poll tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate poll tables; %w", err)
   }

   // 16) Betting / lines
   if err := db.AutoMigrate(
      BettingGame{},
      GameLine{},
   ); err != nil {
      slog.Error("could not auto-migrate betting tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate betting tables; %w", err)
   }

   // 17) Draft
   if err := db.AutoMigrate(
      DraftTeam{},
      DraftPosition{},
      DraftPickHometownInfo{},
      DraftPick{},
   ); err != nil {
      slog.Error("could not auto-migrate draft tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate draft tables; %w", err)
   }

   // 18) Coaches
   if err := db.AutoMigrate(
      Coach{},
      CoachSeason{},
   ); err != nil {
      slog.Error("could not auto-migrate coach tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate coach tables; %w", err)
   }

   // 19) WEPA / metrics
   if err := db.AutoMigrate(
      AdjustedTeamMetrics{},
      PlayerWeightedEPA{},
      KickerPAAR{},
      TeamATS{},
      TeamTalent{},
      GameHavocStatSide{},
      GameHavocStats{},
      AdvancedRateMetrics{},
      AdvancedHavoc{},
      AdvancedFieldPosition{},
      AdvancedSeasonStatSide{},
      AdvancedSeasonStat{},
      AdvancedGameStatSide{},
      AdvancedGameStat{},
   ); err != nil {
      slog.Error("could not auto-migrate metrics tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate metrics tables; %w", err)
   }

   // 20) Misc
   if err := db.AutoMigrate(
      UserInfo{},
      Int32List{},
   ); err != nil {
      slog.Error("could not auto-migrate misc tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate misc tables; %w", err)
   }

   return nil
}

// IsInitialized returns true if the DB appears initialized.
func (db *Database) IsInitialized() (bool, error) {
   type existsRow struct {
      Exists bool
   }

   // 1) schema exists?
   var schema existsRow
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

   // 2) sentinel tables exist?
   // Pick tables that are created across the Initialize() phases so we can
   // detect partial/failed initialization.
   requiredTables := []string{
      // reference/dims
      "venues",
      "conferences",
      "teams",

      // spine
      "games",

      // plays/drives
      "drives",
      "plays",
      "play_types",
      "play_stat_types",
      "play_stats",

      // nested game stats
      "game_team_stats",
      "game_player_stats",

      // other groups
      "recruits",
      "team_sp",
      "poll_weeks",
      "betting_games",
      "draft_picks",
      "coaches",

      // “late” misc
      "int32_lists",
   }

   var foundCount int64
   if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'cfbd'
		  AND table_name IN ?;
	`, requiredTables).Scan(&foundCount).Error; err != nil {
      slog.Error("could not check for sentinel tables", "err", err.Error())
      return false, fmt.Errorf("could not check for sentinel tables; %w", err)
   }

   if foundCount != int64(len(requiredTables)) {
      return false, nil
   }

   return true, nil
}

func (db *Database) InsertConferences(ctx context.Context, conferences []*cfbd.Conference) error {
   if len(conferences) == 0 {
      return nil
   }

   models := make([]Conference, 0, len(conferences))
   for _, c := range conferences {
      if c == nil {
         continue
      }
      id := c.GetId()
      if id == 0 {
         continue
      }

      models = append(models, Conference{
         ID:             int32(id),
         Name:           strings.TrimSpace(c.GetName()),
         ShortName:      strings.TrimSpace(c.GetShortName()),
         Abbreviation:   strings.TrimSpace(c.GetAbbreviation()),
         Classification: strings.TrimSpace(c.GetClassification()),
      })
   }

   if len(models) == 0 {
      return nil
   }

   if err := db.WithContext(ctx).
      Clauses(clause.OnConflict{
         Columns: []clause.Column{{Name: "id"}},
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

func (db *Database) InsertVenues(ctx context.Context, venues []*cfbd.Venue) error {
   if len(venues) == 0 {
      return nil
   }

   models := make([]Venue, 0, len(venues))
   for _, v := range venues {
      if v == nil {
         continue
      }

      // Venue ID is NOT optional per your note.
      id := v.GetId()
      if id == 0 {
         continue
      }

      // For proto3 optional scalars, the generated struct contains pointer fields
      // (e.g. v.Latitude != nil). We avoid relying on getters for presence.
      var lat *float64
      if v.Latitude != nil {
         x := *v.Latitude
         lat = &x
      }
      var lon *float64
      if v.Longitude != nil {
         x := *v.Longitude
         lon = &x
      }
      var cap *int32
      if v.Capacity != nil {
         x := *v.Capacity
         cap = &x
      }
      var cy *int32
      if v.ConstructionYear != nil {
         x := *v.ConstructionYear
         cy = &x
      }
      var grass *bool
      if v.Grass != nil {
         x := *v.Grass
         grass = &x
      }
      var dome *bool
      if v.Dome != nil {
         x := *v.Dome
         dome = &x
      }

      models = append(models, Venue{
         ID:               int32(id),
         Name:             strings.TrimSpace(v.GetName()),
         City:             strings.TrimSpace(v.GetCity()),
         State:            strings.TrimSpace(v.GetState()),
         Zip:              strings.TrimSpace(v.GetZip()),
         CountryCode:      strings.TrimSpace(v.GetCountryCode()),
         Timezone:         strings.TrimSpace(v.GetTimezone()),
         Latitude:         lat,
         Longitude:        lon,
         Elevation:        strings.TrimSpace(v.GetElevation()),
         Capacity:         cap,
         ConstructionYear: cy,
         Grass:            grass,
         Dome:             dome,
      })
   }

   if len(models) == 0 {
      return nil
   }

   if err := db.WithContext(ctx).
      Clauses(clause.OnConflict{
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
      }).
      CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not upsert venues", "err", err.Error())
      return fmt.Errorf("could not upsert venues; %w", err)
   }

   return nil
}

func (db *Database) InsertPlayTypes(ctx context.Context, playTypes []*cfbd.PlayType) error {
   if len(playTypes) == 0 {
      return nil
   }

   models := make([]PlayType, 0, len(playTypes))
   for _, pt := range playTypes {
      if pt == nil {
         continue
      }
      id := pt.GetId()
      if id == 0 {
         continue
      }
      models = append(models, PlayType{
         ID:           int32(id),
         Text:         strings.TrimSpace(pt.GetText()),
         Abbreviation: strings.TrimSpace(pt.GetAbbreviation()),
      })
   }

   if len(models) == 0 {
      return nil
   }

   if err := db.WithContext(ctx).
      Clauses(clause.OnConflict{
         Columns: []clause.Column{{Name: "id"}},
         DoUpdates: clause.AssignmentColumns([]string{
            "text",
            "abbreviation",
         }),
      }).
      CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not upsert play types", "err", err.Error())
      return fmt.Errorf("could not upsert play types; %w", err)
   }

   return nil
}

func (db *Database) InsertPlayStatTypes(ctx context.Context, names []string) error {
   // Normalize + dedupe
   uniq := make(map[string]struct{}, len(names))
   clean := make([]string, 0, len(names))
   for _, n := range names {
      s := strings.TrimSpace(n)
      if s == "" {
         continue
      }
      if _, ok := uniq[s]; ok {
         continue
      }
      uniq[s] = struct{}{}
      clean = append(clean, s)
   }
   if len(clean) == 0 {
      return nil
   }

   // Assign IDs deterministically in this batch (1..N).
   // If you already have rows in cfbd.play_stat_types, this will conflict.
   models := make([]PlayStatType, 0, len(clean))
   for i, name := range clean {
      models = append(models, PlayStatType{
         ID:   int32(i + 1),
         Name: name,
      })
   }

   if err := db.WithContext(ctx).CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not insert play stat types", "err", err.Error())
      return fmt.Errorf("could not insert play stat types; %w", err)
   }

   return nil
}

func (db *Database) InsertDraftTeams(ctx context.Context, teams []*cfbd.DraftTeam) error {
   if len(teams) == 0 {
      return nil
   }

   // DraftTeam in your model uses an auto-increment PK; API provides no ID.
   // We'll insert best-effort and use ON CONFLICT DO NOTHING (no target).
   models := make([]DraftTeam, 0, len(teams))
   for _, t := range teams {
      if t == nil {
         continue
      }
      location := strings.TrimSpace(t.GetLocation())
      if location == "" {
         continue
      }
      models = append(models, DraftTeam{
         Location:    location,
         Nickname:    strings.TrimSpace(t.GetNickname()),
         DisplayName: strings.TrimSpace(t.GetDisplayName()),
         Logo:        strings.TrimSpace(t.GetLogo()),
      })
   }

   if len(models) == 0 {
      return nil
   }

   if err := db.WithContext(ctx).
      Clauses(clause.OnConflict{DoNothing: true}).
      CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not insert draft teams", "err", err.Error())
      return fmt.Errorf("could not insert draft teams; %w", err)
   }

   return nil
}

func (db *Database) InsertDraftPositions(ctx context.Context, positions []*cfbd.DraftPosition) error {
   if len(positions) == 0 {
      return nil
   }

   // DraftPosition in your model uses an auto-increment PK; API provides no ID.
   models := make([]DraftPosition, 0, len(positions))
   for _, p := range positions {
      if p == nil {
         continue
      }
      name := strings.TrimSpace(p.GetName())
      abbr := strings.TrimSpace(p.GetAbbreviation())
      if name == "" && abbr == "" {
         continue
      }
      models = append(models, DraftPosition{
         Name:         name,
         Abbreviation: abbr,
      })
   }

   if len(models) == 0 {
      return nil
   }

   if err := db.WithContext(ctx).
      Clauses(clause.OnConflict{DoNothing: true}).
      CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not insert draft positions", "err", err.Error())
      return fmt.Errorf("could not insert draft positions; %w", err)
   }

   return nil
}

func (db *Database) InsertFieldGoalEP(
   ctx context.Context,
   items []*cfbd.FieldGoalEP,
) error {
   if len(items) == 0 {
      return nil
   }

   models := make([]FieldGoalEP, 0, len(items))
   for _, it := range items {
      if it == nil {
         continue
      }
      // Composite PK in model: (yards_to_goal, distance)
      models = append(models, FieldGoalEP{
         YardsToGoal:    int32(it.GetYardsToGoal()),
         Distance:       int32(it.GetDistance()),
         ExpectedPoints: it.GetExpectedPoints(),
      })
   }

   if len(models) == 0 {
      return nil
   }

   if err := db.WithContext(ctx).
      Clauses(clause.OnConflict{
         Columns: []clause.Column{
            {Name: "yards_to_goal"},
            {Name: "distance"},
         },
         DoUpdates: clause.AssignmentColumns([]string{
            "expected_points",
         }),
      }).
      CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not upsert field goal EP", "err", err.Error())
      return fmt.Errorf("could not upsert field goal EP; %w", err)
   }

   return nil
}

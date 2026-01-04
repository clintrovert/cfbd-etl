package db

import (
   "context"
   "errors"
   "fmt"
   "log/slog"
   "strings"
   "time"

   "github.com/clintrovert/cfbd-go/cfbd"
   "github.com/lib/pq"
   "gorm.io/driver/postgres"
   "gorm.io/gorm"
   "gorm.io/gorm/clause"
   "gorm.io/gorm/logger"
)

// ErrDsnMissing todo:describe.
var ErrDsnMissing = errors.New("dsn is required")

// Config todo:describe
type Config struct {
   DSN                      string
   MaxOpenConnections       int
   MaxIdleConnections       int
   MaxConnectionLifetimeMin int
}

// Database todo:describe
type Database struct {
   *gorm.DB
}

// NewDatabase todo:describe
func NewDatabase(conf Config) (*Database, error) {
   if strings.TrimSpace(conf.DSN) == "" {
      slog.Error("dsn not provided")
      return nil, ErrDsnMissing
   }

   // Append search_path to DSN if not already present
   dsn := conf.DSN
   if !strings.Contains(dsn, "search_path") {
      separator := "?"
      if strings.Contains(dsn, "?") {
         separator = "&"
      }
      dsn = dsn + separator + "search_path=cfbd,public"
   }

   gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
      Logger:                                   logger.Default.LogMode(logger.Info),
      DisableForeignKeyConstraintWhenMigrating: true,
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
      &Venue{},
      &Conference{},
      &Team{},
   ); err != nil {
      slog.Error("could not auto-migrate reference tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate reference tables; %w", err)
   }

   // 2) Core spine
   if err := db.AutoMigrate(
      &Game{},
   ); err != nil {
      slog.Error("could not auto-migrate games table", "err", err.Error())
      return fmt.Errorf("could not auto-migrate games table; %w", err)
   }

   // 3) Matchups
   if err := db.AutoMigrate(
      &Matchup{},
      &MatchupGame{},
   ); err != nil {
      slog.Error("could not auto-migrate matchup tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate matchup tables; %w", err)
   }

   // 4) Calendar / scoreboard / records
   if err := db.AutoMigrate(
      &CalendarWeek{},
      &Scoreboard{},
      &TeamRecords{},
   ); err != nil {
      slog.Error("could not auto-migrate cal/score tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate cal/score tables; %w", err)
   }

   // 5) Plays / drives + lookup tables
   if err := db.AutoMigrate(
      &PlayType{},
      &PlayStatType{},
      &Drive{},
      &Play{},
      &PlayStat{},
   ); err != nil {
      slog.Error("could not auto-migrate play/drive tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate play/drive tables; %w", err)
   }

   // 6) Game box score stats (nested)
   if err := db.AutoMigrate(
      &GameTeamStats{},
      &GameTeamStatsTeam{},
      &GameTeamStatsTeamStat{},

      &GamePlayerStats{},
      &GamePlayerStatsTeam{},
      &GamePlayerStatCategories{},
      &GamePlayerStatTypes{},
      &GamePlayerStatPlayer{},
   ); err != nil {
      slog.Error("could not auto-migrate game stats tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate game stats tables; %w", err)
   }

   // 7) Live game (nested)
   if err := db.AutoMigrate(
      &LiveGame{},
      &LiveGameTeam{},
      &LiveGameDrive{},
      &LiveGamePlay{},
   ); err != nil {
      slog.Error("could not auto-migrate live game tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate live game tables; %w", err)
   }

   // 8) Media & weather
   if err := db.AutoMigrate(
      &GameMedia{},
      &GameWeather{},
   ); err != nil {
      slog.Error("could not migrate media/weather tables", "err", err.Error())
      return fmt.Errorf("could not migrate media/weather tables; %w", err)
   }

   // 9) Win probability
   if err := db.AutoMigrate(
      &PlayWinProbability{},
      &PregameWinProbability{},
      &FieldGoalEP{},
   ); err != nil {
      slog.Error("could not auto-migrate win prob tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate win prob tables; %w", err)
   }

   // 10) PPA / predicted points
   if err := db.AutoMigrate(
      &PredictedPointsValue{},
      &TeamSeasonPredictedPointsAdded{},
      &TeamGamePredictedPointsAdded{},
      &PlayerGamePredictedPointsAdded{},
      &PlayerSeasonPredictedPointsAdded{},
   ); err != nil {
      slog.Error("could not auto-migrate PPA tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate PPA tables; %w", err)
   }

   // 11) Advanced box score payload table (jsonb)
   if err := db.AutoMigrate(
      &AdvancedBoxScore{},
   ); err != nil {
      slog.Error("could not auto-migrate adv score tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate adv score tables; %w", err)
   }

   // 12) Players / roster / usage / transfers / search
   if err := db.AutoMigrate(
      &RosterPlayer{},
      &PlayerSearchResult{},
      &PlayerUsageSplits{},
      &PlayerUsage{},
      &ReturningProduction{},
      &PlayerTransfer{},
      &PlayerStat{},
      &TeamStat{},
   ); err != nil {
      slog.Error("could not auto-migrate player tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate player tables; %w", err)
   }

   // 13) Recruiting
   if err := db.AutoMigrate(
      &RecruitHometownInfo{},
      &Recruit{},
      &TeamRecruitingRanking{},
      &AggregatedTeamRecruiting{},
   ); err != nil {
      slog.Error("could not auto-migrate recruiting tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate recruiting tables; %w", err)
   }

   // 14) Ratings
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

   // 15) Polls / rankings
   if err := db.AutoMigrate(
      &PollWeek{},
      &Poll{},
      &PollRank{},
   ); err != nil {
      slog.Error("could not auto-migrate poll tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate poll tables; %w", err)
   }

   // 16) Betting / lines
   if err := db.AutoMigrate(
      &BettingGame{},
      &GameLine{},
   ); err != nil {
      slog.Error("could not auto-migrate betting tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate betting tables; %w", err)
   }

   // 17) Draft
   if err := db.AutoMigrate(
      &DraftTeam{},
      &DraftPosition{},
      &DraftPickHometownInfo{},
      &DraftPick{},
   ); err != nil {
      slog.Error("could not auto-migrate draft tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate draft tables; %w", err)
   }

   // 18) Coaches
   if err := db.AutoMigrate(
      &Coach{},
      &CoachSeason{},
   ); err != nil {
      slog.Error("could not auto-migrate coach tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate coach tables; %w", err)
   }

   // 19) WEPA / metrics
   if err := db.AutoMigrate(
      &AdjustedTeamMetrics{},
      &PlayerWeightedEPA{},
      &KickerPAAR{},
      &TeamATS{},
      &TeamTalent{},
      &GameHavocStatSide{},
      &GameHavocStats{},
      &AdvancedRateMetrics{},
      &AdvancedHavoc{},
      &AdvancedFieldPosition{},
      &AdvancedSeasonStatSide{},
      &AdvancedSeasonStat{},
      &AdvancedGameStatSide{},
      &AdvancedGameStat{},
   ); err != nil {
      slog.Error("could not auto-migrate metrics tables", "err", err.Error())
      return fmt.Errorf("could not auto-migrate metrics tables; %w", err)
   }

   // 20) Misc
   if err := db.AutoMigrate(
      &UserInfo{},
      &Int32List{},
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

// InsertConferences todo:describe.
func (db *Database) InsertConferences(
   ctx context.Context,
   conferences []*cfbd.Conference,
) error {
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
         ID:             id,
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

// InsertVenues todo:describe.
func (db *Database) InsertVenues(
   ctx context.Context,
   venues []*cfbd.Venue,
) error {
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

      // For proto3 optional scalars, the generated struct contains
      //  pointer fields (e.g. v.Latitude != nil).
      //  We avoid relying on getters for presence.
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
      var capacity *int32
      if v.Capacity != nil {
         x := *v.Capacity
         capacity = &x
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
         ID:               id,
         Name:             strings.TrimSpace(v.GetName()),
         City:             strings.TrimSpace(v.GetCity()),
         State:            strings.TrimSpace(v.GetState()),
         Zip:              strings.TrimSpace(v.GetZip()),
         CountryCode:      strings.TrimSpace(v.GetCountryCode()),
         Timezone:         strings.TrimSpace(v.GetTimezone()),
         Latitude:         lat,
         Longitude:        lon,
         Elevation:        strings.TrimSpace(v.GetElevation()),
         Capacity:         capacity,
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

// InsertPlayTypes todo:describe.
func (db *Database) InsertPlayTypes(
   ctx context.Context,
   playTypes []*cfbd.PlayType,
) error {
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
         ID:           id,
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

// InsertPlayStatTypes todo:describe.
func (db *Database) InsertPlayStatTypes(
   ctx context.Context,
   names []string,
) error {
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
   // We assume these stat types will not change with much frequency.
   models := make([]PlayStatType, 0, len(clean))
   for i, name := range clean {
      models = append(models, PlayStatType{
         ID:   int32(i + 1),
         Name: name,
      })
   }

   if err := db.WithContext(ctx).
      CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not insert play stat types", "err", err.Error())
      return fmt.Errorf("could not insert play stat types; %w", err)
   }

   return nil
}

// InsertDraftTeams todo:describe.
func (db *Database) InsertDraftTeams(
   ctx context.Context,
   teams []*cfbd.DraftTeam,
) error {
   if len(teams) == 0 {
      return nil
   }

   // DraftTeam in model uses an auto-increment PK; API provides no ID.
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

// InsertDraftPositions todo:describe.
func (db *Database) InsertDraftPositions(
   ctx context.Context,
   positions []*cfbd.DraftPosition,
) error {
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

// InsertFieldGoalEP todo:describe.
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
         YardsToGoal:    it.GetYardsToGoal(),
         Distance:       it.GetDistance(),
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

// InsertTeams inserts Team rows into cfbd.teams.
// Assumes Venues table exists and is populated.
func (db *Database) InsertTeams(ctx context.Context, teams []*cfbd.Team) error {
   if len(teams) == 0 {
      return nil
   }

   // De-dupe by team id
   byID := make(map[int32]Team, len(teams))

   for _, t := range teams {
      if t == nil {
         continue
      }

      id := int32(t.GetId())
      if id == 0 {
         continue
      }

      var venueID *int32
      if loc := t.GetLocation(); loc != nil {
         // venue id is NOT optional (per your note)
         vid := int32(loc.GetId())
         if vid != 0 {
            venueID = &vid
         }
      }

      byID[id] = Team{
         ID:             id,
         School:         strings.TrimSpace(t.GetSchool()),
         Mascot:         strings.TrimSpace(t.GetMascot()),
         Abbreviation:   strings.TrimSpace(t.GetAbbreviation()),
         AlternateNames: toStringArray(t.GetAlternateNames()),
         Conference:     strings.TrimSpace(t.GetConference()),
         Division:       strings.TrimSpace(t.GetDivision()),
         Classification: strings.TrimSpace(t.GetClassification()),
         Color:          strings.TrimSpace(t.GetColor()),
         AlternateColor: strings.TrimSpace(t.GetAlternateColor()),
         Logos:          toStringArray(t.GetLogos()),
         Twitter:        strings.TrimSpace(t.GetTwitter()),
         VenueID:        venueID,
      }
   }

   if len(byID) == 0 {
      return nil
   }

   models := make([]Team, 0, len(byID))
   for _, m := range byID {
      // school is effectively required for a useful team row
      if m.School == "" {
         continue
      }
      models = append(models, m)
   }

   if len(models) == 0 {
      return nil
   }

   if err := db.WithContext(ctx).
      Clauses(clause.OnConflict{
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
      }).
      CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not upsert teams", "err", err.Error())
      return fmt.Errorf("could not upsert teams; %w", err)
   }

   return nil
}

// InsertCalendarWeeks todo:describe
func (db *Database) InsertCalendarWeeks(
   ctx context.Context,
   weeks []*cfbd.CalendarWeek,
) error {
   if len(weeks) == 0 {
      return nil
   }

   models := make([]CalendarWeek, 0, len(weeks))
   for _, w := range weeks {
      if w == nil {
         continue
      }
      season := w.GetSeason()
      week := w.GetWeek()
      seasonType := strings.TrimSpace(w.GetSeasonType())
      if season == 0 || week == 0 || seasonType == "" {
         continue
      }

      var startDate *time.Time
      if w.GetStartDate() != nil {
         t := w.GetStartDate().AsTime()
         startDate = &t
      }
      var endDate *time.Time
      if w.GetEndDate() != nil {
         t := w.GetEndDate().AsTime()
         endDate = &t
      }
      var firstGameStart *time.Time
      if w.GetFirstGameStart() != nil {
         t := w.GetFirstGameStart().AsTime()
         firstGameStart = &t
      }
      var lastGameStart *time.Time
      if w.GetLastGameStart() != nil {
         t := w.GetLastGameStart().AsTime()
         lastGameStart = &t
      }

      models = append(models, CalendarWeek{
         Season:         season,
         Week:           week,
         SeasonType:     seasonType,
         StartDate:      startDate,
         EndDate:        endDate,
         FirstGameStart: firstGameStart,
         LastGameStart:  lastGameStart,
      })
   }

   if len(models) == 0 {
      return nil
   }

   if err := db.WithContext(ctx).
      Clauses(clause.OnConflict{
         Columns: []clause.Column{
            {Name: "season"},
            {Name: "week"},
            {Name: "season_type"},
         },
         DoUpdates: clause.AssignmentColumns([]string{
            "start_date",
            "end_date",
            "first_game_start",
            "last_game_start",
         }),
      }).
      CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not upsert calendar weeks", "err", err.Error())
      return fmt.Errorf("could not upsert calendar weeks; %w", err)
   }

   return nil
}

func (db *Database) InsertGames(ctx context.Context, games []*cfbd.Game) error {
   if len(games) == 0 {
      return nil
   }

   models := make([]Game, 0, len(games))
   for _, g := range games {
      if g == nil {
         continue
      }

      id := g.GetId()
      if id == 0 {
         continue
      }

      var startDate *time.Time
      if g.GetStartDate() != nil {
         t := g.GetStartDate().AsTime()
         startDate = &t
      }

      // Optional scalars in proto3 => presence via exported pointer fields
      var attendance *int32
      if g.Attendance != nil {
         x := *g.Attendance
         attendance = &x
      }

      var venueID *int32
      if g.VenueId != nil {
         x := *g.VenueId
         venueID = &x
      }

      var homeID *int32
      if g.HomeId != nil {
         x := *g.HomeId
         homeID = &x
      }
      var homePoints *int32
      if g.HomePoints != nil {
         x := *g.HomePoints
         homePoints = &x
      }

      var awayID *int32
      if g.AwayId != nil {
         x := *g.AwayId
         awayID = &x
      }
      var awayPoints *int32
      if g.AwayPoints != nil {
         x := *g.AwayPoints
         awayPoints = &x
      }

      var homePostWinProb *float64
      if g.HomePostgameWinProbability != nil {
         x := *g.HomePostgameWinProbability
         homePostWinProb = &x
      }
      var awayPostWinProb *float64
      if g.AwayPostgameWinProbability != nil {
         x := *g.AwayPostgameWinProbability
         awayPostWinProb = &x
      }

      var homePregameElo *int32
      if g.HomePregameElo != nil {
         x := *g.HomePregameElo
         homePregameElo = &x
      }
      var homePostgameElo *int32
      if g.HomePostgameElo != nil {
         x := *g.HomePostgameElo
         homePostgameElo = &x
      }
      var awayPregameElo *int32
      if g.AwayPregameElo != nil {
         x := *g.AwayPregameElo
         awayPregameElo = &x
      }
      var awayPostgameElo *int32
      if g.AwayPostgameElo != nil {
         x := *g.AwayPostgameElo
         awayPostgameElo = &x
      }

      var excitementIndex *float64
      if g.ExcitementIndex != nil {
         x := *g.ExcitementIndex
         excitementIndex = &x
      }

      models = append(models, Game{
         ID:                     id,
         Season:                 g.GetSeason(),
         Week:                   g.GetWeek(),
         SeasonType:             strings.TrimSpace(g.GetSeasonType()),
         StartDate:              startDate,
         StartTimeTBD:           g.GetStartTime_TBD(),
         Completed:              g.GetCompleted(),
         NeutralSite:            g.GetNeutralSite(),
         ConferenceGame:         g.GetConferenceGame(),
         Attendance:             attendance,
         VenueID:                venueID,
         Venue:                  strings.TrimSpace(g.GetVenue()),
         HomeID:                 homeID,
         HomeTeam:               strings.TrimSpace(g.GetHomeTeam()),
         HomeConference:         strings.TrimSpace(g.GetHomeConference()),
         HomeClassification:     strings.TrimSpace(g.GetHomeClassification()),
         HomePoints:             homePoints,
         HomeLineScores:         int32SliceToInt64Array(g.GetHomeLineScores()),
         HomePostWinProbability: homePostWinProb,
         HomePregameElo:         homePregameElo,
         HomePostgameElo:        homePostgameElo,
         AwayID:                 awayID,
         AwayTeam:               strings.TrimSpace(g.GetAwayTeam()),
         AwayConference:         strings.TrimSpace(g.GetAwayConference()),
         AwayClassification:     strings.TrimSpace(g.GetAwayClassification()),
         AwayPoints:             awayPoints,
         AwayLineScores:         int32SliceToInt64Array(g.GetAwayLineScores()),
         AwayPostWinProbability: awayPostWinProb,
         AwayPregameElo:         awayPregameElo,
         AwayPostgameElo:        awayPostgameElo,
         ExcitementIndex:        excitementIndex,
         Highlights:             strings.TrimSpace(g.GetHighlights()),
         Notes:                  strings.TrimSpace(g.GetNotes()),
      })
   }

   if len(models) == 0 {
      return nil
   }

   if err := db.WithContext(ctx).
      Clauses(clause.OnConflict{
         Columns: []clause.Column{{Name: "id"}},
         DoUpdates: clause.AssignmentColumns([]string{
            "season",
            "week",
            "season_type",
            "start_date",
            "start_time_tbd",
            "completed",
            "neutral_site",
            "conference_game",
            "attendance",
            "venue_id",
            "venue",
            "home_id",
            "home_team",
            "home_conference",
            "home_classification",
            "home_points",
            "home_line_scores",
            "home_postgame_win_probability",
            "home_pregame_elo",
            "home_postgame_elo",
            "away_id",
            "away_team",
            "away_conference",
            "away_classification",
            "away_points",
            "away_line_scores",
            "away_postgame_win_probability",
            "away_pregame_elo",
            "away_postgame_elo",
            "excitement_index",
            "highlights",
            "notes",
         }),
      }).
      CreateInBatches(models, 500).Error; err != nil {
      slog.Error("could not upsert games", "err", err.Error())
      return fmt.Errorf("could not upsert games; %w", err)
   }

   return nil
}

func toStringArray(in []string) pq.StringArray {
   if len(in) == 0 {
      // store empty array rather than NULL
      return pq.StringArray{}
   }
   out := make([]string, 0, len(in))
   for _, s := range in {
      v := strings.TrimSpace(s)
      if v == "" {
         continue
      }
      out = append(out, v)
   }
   return pq.StringArray(out)
}

func int32SliceToInt64Array(xs []int32) pq.Int64Array {
   if len(xs) == 0 {
      return nil
   }
   out := make(pq.Int64Array, 0, len(xs))
   for _, v := range xs {
      out = append(out, int64(v))
   }
   return out
}

package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/clintrovert/cfbd-etl/seeder/internal/utils"
	"github.com/clintrovert/cfbd-go/cfbd"
	"gorm.io/datatypes"
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
		Logger: logger.Default.LogMode(
			logger.Info,
		),
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
// InsertTeams inserts Team rows into cfbd.teams.
// Assumes Venues table exists and is populated.
func (db *Database) InsertTeams(
	ctx context.Context,
	teams []*cfbd.Team,
) error {
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
			AlternateNames: utils.ToStringArray(t.GetAlternateNames()),
			Conference:     strings.TrimSpace(t.GetConference()),
			Division:       strings.TrimSpace(t.GetDivision()),
			Classification: strings.TrimSpace(t.GetClassification()),
			Color:          strings.TrimSpace(t.GetColor()),
			AlternateColor: strings.TrimSpace(t.GetAlternateColor()),
			Logos:          utils.ToStringArray(t.GetLogos()),
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

func (db *Database) InsertGames(
	ctx context.Context,
	games []*cfbd.Game,
) error {
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
			HomeLineScores:         utils.Int32SliceToInt64Array(g.GetHomeLineScores()),
			HomePostWinProbability: homePostWinProb,
			HomePregameElo:         homePregameElo,
			HomePostgameElo:        homePostgameElo,
			AwayID:                 awayID,
			AwayTeam:               strings.TrimSpace(g.GetAwayTeam()),
			AwayConference:         strings.TrimSpace(g.GetAwayConference()),
			AwayClassification:     strings.TrimSpace(g.GetAwayClassification()),
			AwayPoints:             awayPoints,
			AwayLineScores:         utils.Int32SliceToInt64Array(g.GetAwayLineScores()),
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

func (db *Database) InsertPlays(
	ctx context.Context,
	plays []*cfbd.Play,
) error {
	if len(plays) == 0 {
		return nil
	}

	models := make([]Play, 0, len(plays))
	for _, p := range plays {
		if p == nil {
			continue
		}

		id := p.GetId()
		if id == "" {
			continue
		}

		var driveNumber *int32
		if p.DriveNumber != nil {
			x := *p.DriveNumber
			driveNumber = &x
		}

		var playNumber *int32
		if p.PlayNumber != nil {
			x := *p.PlayNumber
			playNumber = &x
		}

		var clockMinutes *int32
		if p.Clock != nil {
			if p.Clock.Minutes != nil {
				x := *p.Clock.Minutes
				clockMinutes = &x
			}
		}

		var clockSeconds *int32
		if p.Clock != nil {
			if p.Clock.Seconds != nil {
				x := *p.Clock.Seconds
				clockSeconds = &x
			}
		}

		var offenseTimeouts *int32
		if p.OffenseTimeouts != nil {
			x := *p.OffenseTimeouts
			offenseTimeouts = &x
		}

		var defenseTimeouts *int32
		if p.DefenseTimeouts != nil {
			x := *p.DefenseTimeouts
			defenseTimeouts = &x
		}

		var ppa *float64
		if p.Ppa != nil {
			x := *p.Ppa
			ppa = &x
		}

		models = append(models, Play{
			ID:                id,
			DriveID:           strings.TrimSpace(p.GetDriveId()),
			GameID:            p.GetGameId(),
			DriveNumber:       driveNumber,
			PlayNumber:        playNumber,
			Offense:           strings.TrimSpace(p.GetOffense()),
			OffenseConference: strings.TrimSpace(p.GetOffenseConference()),
			OffenseScore:      p.GetOffenseScore(),
			Defense:           strings.TrimSpace(p.GetDefense()),
			Home:              strings.TrimSpace(p.GetHome()),
			Away:              strings.TrimSpace(p.GetAway()),
			DefenseConference: strings.TrimSpace(p.GetDefenseConference()),
			DefenseScore:      p.GetDefenseScore(),
			Period:            p.GetPeriod(),
			ClockMinutes:      clockMinutes,
			ClockSeconds:      clockSeconds,
			OffenseTimeouts:   offenseTimeouts,
			DefenseTimeouts:   defenseTimeouts,
			Yardline:          p.GetYardline(),
			YardsToGoal:       p.GetYardsToGoal(),
			Down:              p.GetDown(),
			Distance:          p.GetDistance(),
			YardsGained:       p.GetYardsGained(),
			Scoring:           p.GetScoring(),
			PlayType:          strings.TrimSpace(p.GetPlayType()),
			PlayText:          strings.TrimSpace(p.GetPlayText()),
			PPA:               ppa,
			Wallclock:         strings.TrimSpace(p.GetWallclock()),
		})
	}

	if len(models) == 0 {
		return nil
	}

	if err := db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"drive_id",
				"game_id",
				"drive_number",
				"play_number",
				"offense",
				"offense_conference",
				"offense_score",
				"defense",
				"home",
				"away",
				"defense_conference",
				"defense_score",
				"period",
				"clock_minutes",
				"clock_seconds",
				"offense_timeouts",
				"defense_timeouts",
				"yardline",
				"yards_to_goal",
				"down",
				"distance",
				"yards_gained",
				"scoring",
				"play_type",
				"play_text",
				"ppa",
				"wallclock",
			}),
		}).
		CreateInBatches(models, 500).Error; err != nil {
		slog.Error("could not upsert plays", "err", err.Error())
		return fmt.Errorf("could not upsert plays; %w", err)
	}

	return nil
}

func (db *Database) InsertDrives(
	ctx context.Context,
	drives []*cfbd.Drive,
) error {
	if len(drives) == 0 {
		return nil
	}

	models := make([]Drive, 0, len(drives))
	for _, d := range drives {
		if d == nil {
			continue
		}

		id := d.GetId()
		if id == "" {
			continue
		}

		var driveNumber *int32
		if d.DriveNumber != nil {
			x := *d.DriveNumber
			driveNumber = &x
		}

		var startTimeMinutes *int32
		if d.StartTime != nil {
			if d.StartTime.Minutes != nil {
				x := *d.StartTime.Minutes
				startTimeMinutes = &x
			}
		}

		var startTimeSeconds *int32
		if d.StartTime != nil {
			if d.StartTime.Seconds != nil {
				x := *d.StartTime.Seconds
				startTimeSeconds = &x
			}
		}

		var endTimeMinutes *int32
		if d.EndTime != nil {
			if d.EndTime.Minutes != nil {
				x := *d.EndTime.Minutes
				endTimeMinutes = &x
			}
		}

		var endTimeSeconds *int32
		if d.EndTime != nil {
			if d.EndTime.Seconds != nil {
				x := *d.EndTime.Seconds
				endTimeSeconds = &x
			}
		}

		var elapsedMinutes *int32
		if d.Elapsed != nil {
			if d.Elapsed.Minutes != nil {
				x := *d.Elapsed.Minutes
				elapsedMinutes = &x
			}
		}

		var elapsedSeconds *int32
		if d.Elapsed != nil {
			if d.Elapsed.Seconds != nil {
				x := *d.Elapsed.Seconds
				elapsedSeconds = &x
			}
		}

		models = append(models, Drive{
			ID:                id,
			GameID:            d.GetGameId(),
			Offense:           strings.TrimSpace(d.GetOffense()),
			OffenseConference: strings.TrimSpace(d.GetOffenseConference()),
			Defense:           strings.TrimSpace(d.GetDefense()),
			DefenseConference: strings.TrimSpace(d.GetDefenseConference()),
			DriveNumber:       driveNumber,
			Scoring:           d.GetScoring(),
			StartPeriod:       d.GetStartPeriod(),
			StartYardline:     d.GetStartYardline(),
			StartYardsToGoal:  d.GetStartYardsToGoal(),
			StartTimeMinutes:  startTimeMinutes,
			StartTimeSeconds:  startTimeSeconds,
			EndPeriod:         d.GetEndPeriod(),
			EndYardline:       d.GetEndYardline(),
			EndYardsToGoal:    d.GetEndYardsToGoal(),
			EndTimeMinutes:    endTimeMinutes,
			EndTimeSeconds:    endTimeSeconds,
			ElapsedMinutes:    elapsedMinutes,
			ElapsedSeconds:    elapsedSeconds,
			Plays:             d.GetPlays(),
			Yards:             d.GetYards(),
			DriveResult:       strings.TrimSpace(d.GetDriveResult()),
			IsHomeOffense:     d.GetIsHomeOffense(),
			StartOffenseScore: d.GetStartOffenseScore(),
			StartDefenseScore: d.GetStartDefenseScore(),
			EndOffenseScore:   d.GetEndOffenseScore(),
			EndDefenseScore:   d.GetEndDefenseScore(),
		})
	}

	if len(models) == 0 {
		return nil
	}

	if err := db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"game_id",
				"offense",
				"offense_conference",
				"defense",
				"defense_conference",
				"drive_number",
				"scoring",
				"start_period",
				"start_yardline",
				"start_yards_to_goal",
				"start_time_minutes",
				"start_time_seconds",
				"end_period",
				"end_yardline",
				"end_yards_to_goal",
				"end_time_minutes",
				"end_time_seconds",
				"elapsed_minutes",
				"elapsed_seconds",
				"plays",
				"yards",
				"drive_result",
				"is_home_offense",
				"start_offense_score",
				"start_defense_score",
				"end_offense_score",
				"end_defense_score",
			}),
		}).
		CreateInBatches(models, 500).Error; err != nil {
		slog.Error("could not upsert drives", "err", err.Error())
		return fmt.Errorf("could not upsert drives; %w", err)
	}

	return nil
}

func (db *Database) InsertPlayStats(
	ctx context.Context,
	playStats []*cfbd.PlayStat,
) error {
	if len(playStats) == 0 {
		return nil
	}

	models := make([]PlayStat, 0, len(playStats))
	for _, ps := range playStats {
		if ps == nil {
			continue
		}

		// ID is auto-generated (BIGSERIAL), so we set it to 0
		var clockMinutes *float64
		if ps.Clock != nil {
			if ps.Clock.Minutes != nil {
				x := *ps.Clock.Minutes
				clockMinutes = &x
			}
		}

		var clockSeconds *float64
		if ps.Clock != nil {
			if ps.Clock.Seconds != nil {
				x := *ps.Clock.Seconds
				clockSeconds = &x
			}
		}

		models = append(models, PlayStat{
			ID:            0, // Auto-generated by database
			GameID:        ps.GetGameId(),
			Season:        ps.GetSeason(),
			Week:          ps.GetWeek(),
			Team:          strings.TrimSpace(ps.GetTeam()),
			Conference:    strings.TrimSpace(ps.GetConference()),
			Opponent:      strings.TrimSpace(ps.GetOpponent()),
			TeamScore:     ps.GetTeamScore(),
			OpponentScore: ps.GetOpponentScore(),
			DriveID:       strings.TrimSpace(ps.GetDriveId()),
			PlayID:        strings.TrimSpace(ps.GetPlayId()),
			Period:        ps.GetPeriod(),
			ClockMinutes:  clockMinutes,
			ClockSeconds:  clockSeconds,
			YardsToGoal:   ps.GetYardsToGoal(),
			Down:          ps.GetDown(),
			Distance:      ps.GetDistance(),
			AthleteID:     strings.TrimSpace(ps.GetAthleteId()),
			AthleteName:   strings.TrimSpace(ps.GetAthleteName()),
			StatType:      strings.TrimSpace(ps.GetStatType()),
			Stat:          ps.GetStat(),
		})
	}

	if len(models) == 0 {
		return nil
	}

	// Since ID is auto-generated and there's no unique constraint in the schema,
	// we use DoNothing to avoid errors on potential duplicates
	if err := db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(models, 500).Error; err != nil {
		slog.Error("could not insert play stats", "err", err.Error())
		return fmt.Errorf("could not insert play stats; %w", err)
	}

	return nil
}

// InsertGameWeather inserts game weather data.
func (db *Database) InsertGameWeather(
	ctx context.Context,
	weather []*cfbd.GameWeather,
) error {
	if len(weather) == 0 {
		return nil
	}

	models := make([]GameWeather, 0, len(weather))
	for _, w := range weather {
		if w == nil {
			continue
		}
		var startTime *time.Time
		if w.StartTime != nil {
			t := w.StartTime.AsTime()
			startTime = &t
		}

		venueID := w.VenueId // protobuf field
		models = append(models, GameWeather{
			ID:                   w.Id, // protobuf field
			Season:               w.Season,
			Week:                 w.Week,
			SeasonType:           w.SeasonType,
			StartTime:            startTime,
			GameIndoors:          w.GameIndoors,
			HomeTeam:             w.HomeTeam,
			HomeConference:       w.HomeConference,
			AwayTeam:             w.AwayTeam,
			AwayConference:       w.AwayConference,
			VenueID:              venueID,
			Venue:                w.Venue,
			Temperature:          w.Temperature,
			DewPoint:             w.DewPoint,
			Humidity:             w.Humidity,
			Precipitation:        w.Precipitation,
			Snowfall:             w.Snowfall,
			WindDirection:        w.WindDirection,
			WindSpeed:            w.WindSpeed,
			Pressure:             w.Pressure,
			WeatherConditionCode: w.WeatherConditionCode,
			WeatherCondition:     w.WeatherCondition,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertGameMedia inserts game media data.
func (db *Database) InsertGameMedia(
	ctx context.Context,
	media []*cfbd.GameMedia,
) error {
	if len(media) == 0 {
		return nil
	}

	models := make([]GameMedia, 0, len(media))
	for _, m := range media {
		if m == nil {
			continue
		}
		var startTime *time.Time
		if m.StartTime != nil {
			t := m.StartTime.AsTime()
			startTime = &t
		}

		models = append(models, GameMedia{
			ID:         m.Id, // protobuf field
			Season:     m.Season,
			Week:       m.Week,
			SeasonType: m.SeasonType,
			StartTime:  startTime,
			// Check exact name in doc: IsStartTime_TBD?
			IsStartTimeTBD: m.IsStartTime_TBD,
			HomeTeam:       m.HomeTeam,
			HomeConference: m.HomeConference,
			AwayTeam:       m.AwayTeam,
			AwayConference: m.AwayConference,
			MediaType:      m.MediaType,
			Outlet:         m.Outlet,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertBettingLines inserts game betting lines.
func (db *Database) InsertBettingLines(
	ctx context.Context,
	lines []*cfbd.BettingGame,
) error {
	if len(lines) == 0 {
		return nil
	}

	models := make([]BettingGame, 0, len(lines))
	for _, l := range lines {
		if l == nil {
			continue
		}
		var startDate *time.Time
		if l.StartDate != nil {
			t := l.StartDate.AsTime()
			startDate = &t
		}

		gameLines := make([]GameLine, 0, len(l.Lines))
		for _, gl := range l.Lines {
			if gl == nil {
				continue
			}
			gameLines = append(gameLines, GameLine{
				GameID:          l.Id, // protobuf field
				Provider:        gl.Provider,
				Spread:          gl.Spread,
				FormattedSpread: gl.FormattedSpread,
				SpreadOpen:      gl.SpreadOpen,
				OverUnder:       gl.OverUnder,
				OverUnderOpen:   gl.OverUnderOpen,
				HomeMoneyline:   gl.HomeMoneyline,
				AwayMoneyline:   gl.AwayMoneyline,
			})
		}

		models = append(models, BettingGame{
			ID:                 l.Id, // protobuf field
			Season:             l.Season,
			SeasonType:         l.SeasonType,
			Week:               l.Week,
			StartDate:          startDate,
			HomeTeamID:         l.HomeTeamId, // protobuf field
			HomeTeam:           l.HomeTeam,
			HomeConference:     l.HomeConference,
			HomeClassification: l.HomeClassification,
			HomeScore:          l.HomeScore,
			AwayTeamID:         l.AwayTeamId, // protobuf field
			AwayTeam:           l.AwayTeam,
			AwayConference:     l.AwayConference,
			AwayClassification: l.AwayClassification,
			AwayScore:          l.AwayScore,
			Lines:              gameLines,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertTeamRecords inserts team records.
func (db *Database) InsertTeamRecords(
	ctx context.Context,
	records []*cfbd.TeamRecords,
) error {
	if len(records) == 0 {
		return nil
	}

	models := make([]TeamRecords, 0, len(records))
	for _, r := range records {
		if r == nil {
			continue
		}

		// Helper to safely get total record
		getRec := func(rec *cfbd.TeamRecord) (games, wins, losses, ties int32) {
			if rec == nil {
				return 0, 0, 0, 0
			}
			return rec.Games, rec.Wins, rec.Losses, rec.Ties
		}

		totGames, totWins, totLosses, totTies := getRec(r.Total)
		confGames, confWins, confLosses, confTies := getRec(r.ConferenceGames)
		homeGames, homeWins, homeLosses, homeTies := getRec(r.HomeGames)
		awayGames, awayWins, awayLosses, awayTies := getRec(r.AwayGames)
		neuGames, neuWins, neuLosses, neuTies := getRec(r.NeutralSiteGames)
		regGames, regWins, regLosses, regTies := getRec(r.RegularSeason)
		postGames, postWins, postLosses, postTies := getRec(r.Postseason)

		var teamID int32
		if r.TeamId != nil {
			teamID = *r.TeamId
		}

		models = append(models, TeamRecords{
			Year:                   r.Year,
			Team:                   r.Team,
			TeamID:                 &teamID,
			Classification:         r.Classification,
			Conference:             r.Conference,
			Division:               r.Division,
			ExpectedWins:           r.ExpectedWins,
			TotalGames:             totGames,
			TotalWins:              totWins,
			TotalLosses:            totLosses,
			TotalTies:              totTies,
			ConferenceGamesGames:   confGames,
			ConferenceGamesWins:    confWins,
			ConferenceGamesLosses:  confLosses,
			ConferenceGamesTies:    confTies,
			HomeGamesGames:         homeGames,
			HomeGamesWins:          homeWins,
			HomeGamesLosses:        homeLosses,
			HomeGamesTies:          homeTies,
			AwayGamesGames:         awayGames,
			AwayGamesWins:          awayWins,
			AwayGamesLosses:        awayLosses,
			AwayGamesTies:          awayTies,
			NeutralSiteGamesGames:  neuGames,
			NeutralSiteGamesWins:   neuWins,
			NeutralSiteGamesLosses: neuLosses,
			NeutralSiteGamesTies:   neuTies,
			RegularSeasonGames:     regGames,
			RegularSeasonWins:      regWins,
			RegularSeasonLosses:    regLosses,
			RegularSeasonTies:      regTies,
			PostseasonGames:        postGames,
			PostseasonWins:         postWins,
			PostseasonLosses:       postLosses,
			PostseasonTies:         postTies,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertTeamTalent inserts team talent composite rankings.
func (db *Database) InsertTeamTalent(
	ctx context.Context,
	talent []*cfbd.TeamTalent,
) error {
	if len(talent) == 0 {
		return nil
	}

	models := make([]TeamTalent, 0, len(talent))
	for _, t := range talent {
		if t == nil {
			continue
		}
		models = append(models, TeamTalent{
			Year:   t.Year,
			Team:   t.Team, // protobuf field is Team, not School
			Talent: t.Talent,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertTeamATS inserts team ATS records.
func (db *Database) InsertTeamATS(
	ctx context.Context,
	ats []*cfbd.TeamATS,
) error {
	if len(ats) == 0 {
		return nil
	}

	models := make([]TeamATS, 0, len(ats))
	for _, a := range ats {
		if a == nil {
			continue
		}
		models = append(models, TeamATS{
			Year:           a.Year,
			TeamID:         a.TeamId, // protobuf field
			Team:           a.Team,
			Conference:     a.Conference,
			Games:          a.Games,
			AtsWins:        a.AtsWins,
			AtsLosses:      a.AtsLosses,
			AtsPushes:      a.AtsPushes,
			AvgCoverMargin: a.AvgCoverMargin,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertTeamSP inserts team SP+ ratings.
func (db *Database) InsertTeamSP(
	ctx context.Context,
	ratings []*cfbd.TeamSP,
) error {
	if len(ratings) == 0 {
		return nil
	}

	models := make([]TeamSP, 0, len(ratings))
	for _, r := range ratings {
		if r == nil {
			continue
		}

		payload, err := json.Marshal(r)
		if err != nil {
			slog.Error("failed to marshal team sp payload", "err", err)
			continue
		}

		models = append(models, TeamSP{
			Year:       r.Year,
			Team:       r.Team,
			Conference: r.Conference,
			Payload:    datatypes.JSON(payload),
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertConferenceSP inserts conference SP+ ratings.
func (db *Database) InsertConferenceSP(
	ctx context.Context,
	ratings []*cfbd.ConferenceSP,
) error {
	if len(ratings) == 0 {
		return nil
	}

	models := make([]ConferenceSP, 0, len(ratings))
	for _, r := range ratings {
		if r == nil {
			continue
		}

		payload, err := json.Marshal(r)
		if err != nil {
			slog.Error("failed to marshal conference sp payload", "err", err)
			continue
		}

		models = append(models, ConferenceSP{
			Year:       r.Year,
			Conference: r.Conference,
			Payload:    datatypes.JSON(payload),
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertTeamSRS inserts team SRS ratings.
func (db *Database) InsertTeamSRS(
	ctx context.Context,
	ratings []*cfbd.TeamSRS,
) error {
	if len(ratings) == 0 {
		return nil
	}

	models := make([]TeamSRS, 0, len(ratings))
	for _, r := range ratings {
		if r == nil {
			continue
		}
		models = append(models, TeamSRS{
			Year:       r.Year,
			Team:       r.Team,
			Conference: r.Conference,
			Division:   r.Division,
			Rating:     r.Rating,
			Ranking:    r.Ranking,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertTeamElo inserts team Elo ratings.
func (db *Database) InsertTeamElo(
	ctx context.Context,
	ratings []*cfbd.TeamElo,
) error {
	if len(ratings) == 0 {
		return nil
	}

	models := make([]TeamElo, 0, len(ratings))
	for _, r := range ratings {
		if r == nil {
			continue
		}
		models = append(models, TeamElo{
			Year:       r.Year,
			Team:       r.Team,
			Conference: r.Conference,
			Elo:        r.Elo,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertTeamFPI inserts team FPI ratings.
func (db *Database) InsertTeamFPI(
	ctx context.Context,
	ratings []*cfbd.TeamFPI,
) error {
	if len(ratings) == 0 {
		return nil
	}

	models := make([]TeamFPI, 0, len(ratings))
	for _, r := range ratings {
		if r == nil {
			continue
		}

		payload, err := json.Marshal(r)
		if err != nil {
			slog.Error("failed to marshal team fpi payload", "err", err)
			continue
		}

		models = append(models, TeamFPI{
			Year:       r.Year,
			Team:       r.Team,
			Conference: r.Conference,
			Payload:    datatypes.JSON(payload),
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertAdjustedTeamMetrics inserts adjusted team metrics (WEPA/EPA).
func (db *Database) InsertAdjustedTeamMetrics(
	ctx context.Context,
	metrics []*cfbd.AdjustedTeamMetrics,
) error {
	if len(metrics) == 0 {
		return nil
	}

	models := make([]AdjustedTeamMetrics, 0, len(metrics))
	for _, m := range metrics {
		if m == nil {
			continue
		}

		epaRush, epaPass, epaTotal := 0.0, 0.0, 0.0
		if m.Epa != nil {
			epaRush = m.Epa.Rushing
			epaPass = m.Epa.Passing
			epaTotal = m.Epa.Total
		}

		epaAllowRush, epaAllowPass, epaAllowTotal := 0.0, 0.0, 0.0
		if m.EpaAllowed != nil {
			epaAllowRush = m.EpaAllowed.Rushing
			epaAllowPass = m.EpaAllowed.Passing
			epaAllowTotal = m.EpaAllowed.Total
		}

		srPass, srStd, srTot := 0.0, 0.0, 0.0
		if m.SuccessRate != nil {
			srPass = m.SuccessRate.PassingDowns
			srStd = m.SuccessRate.StandardDowns
			srTot = m.SuccessRate.Total
		}

		srAllowPass, srAllowStd, srAllowTot := 0.0, 0.0, 0.0
		if m.SuccessRateAllowed != nil {
			srAllowPass = m.SuccessRateAllowed.PassingDowns
			srAllowStd = m.SuccessRateAllowed.StandardDowns
			srAllowTot = m.SuccessRateAllowed.Total
		}

		rushHigh, rushOpen, rushSec, rushLine := 0.0, 0.0, 0.0, 0.0
		if m.Rushing != nil {
			rushHigh = m.Rushing.HighlightYards
			rushOpen = m.Rushing.OpenFieldYards
			rushSec = m.Rushing.SecondLevelYards
			rushLine = m.Rushing.LineYards
		}

		rushAllowHigh, rushAllowOpen, rushAllowSec, rushAllowLine := 0.0, 0.0, 0.0, 0.0
		if m.RushingAllowed != nil {
			rushAllowHigh = m.RushingAllowed.HighlightYards
			rushAllowOpen = m.RushingAllowed.OpenFieldYards
			rushAllowSec = m.RushingAllowed.SecondLevelYards
			rushAllowLine = m.RushingAllowed.LineYards
		}

		models = append(models, AdjustedTeamMetrics{
			Year:                            m.Year,
			TeamID:                          m.TeamId, // protobuf field
			Team:                            m.Team,
			Conference:                      m.Conference,
			EpaRushing:                      epaRush,
			EpaPassing:                      epaPass,
			EpaTotal:                        epaTotal,
			EpaAllowedRushing:               epaAllowRush,
			EpaAllowedPassing:               epaAllowPass,
			EpaAllowedTotal:                 epaAllowTotal,
			SuccessRatePassingDowns:         srPass,
			SuccessRateStandardDowns:        srStd,
			SuccessRateTotal:                srTot,
			SuccessRateAllowedPassingDowns:  srAllowPass,
			SuccessRateAllowedStandardDowns: srAllowStd,
			SuccessRateAllowedTotal:         srAllowTot,
			RushingHighlightYards:           rushHigh,
			RushingOpenFieldYards:           rushOpen,
			RushingSecondLevelYards:         rushSec,
			RushingLineYards:                rushLine,
			RushingAllowedHighlightYards:    rushAllowHigh,
			RushingAllowedOpenFieldYards:    rushAllowOpen,
			RushingAllowedSecondLevelYards:  rushAllowSec,
			RushingAllowedLineYards:         rushAllowLine,
			Explosiveness:                   m.Explosiveness,
			ExplosivenessAllowed:            m.ExplosivenessAllowed,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertPlayerWeightedEPA inserts player weighted EPA.
func (db *Database) InsertPlayerWeightedEPA(
	ctx context.Context,
	metrics []*cfbd.PlayerWeightedEPA,
) error {
	if len(metrics) == 0 {
		return nil
	}

	models := make([]PlayerWeightedEPA, 0, len(metrics))
	for _, m := range metrics {
		if m == nil {
			continue
		}
		models = append(models, PlayerWeightedEPA{
			Year:        m.Year,
			AthleteID:   m.AthleteId, // protobuf field
			AthleteName: m.AthleteName,
			Position:    m.Position,
			Team:        m.Team,
			Conference:  m.Conference,
			WEPA:        m.Wepa,
			Plays:       m.Plays,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertKickerPAAR inserts kicker PAAR.
func (db *Database) InsertKickerPAAR(
	ctx context.Context,
	kickers []*cfbd.KickerPAAR,
) error {
	if len(kickers) == 0 {
		return nil
	}

	models := make([]KickerPAAR, 0, len(kickers))
	for _, k := range kickers {
		if k == nil {
			continue
		}
		models = append(models, KickerPAAR{
			Year:        k.Year,
			AthleteID:   k.AthleteId, // protobuf field
			AthleteName: k.AthleteName,
			Team:        k.Team,
			Conference:  k.Conference,
			PAAR:        k.Paar,
			Attempts:    k.Attempts,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertReturningProduction inserts returning production.
func (db *Database) InsertReturningProduction(
	ctx context.Context,
	production []*cfbd.ReturningProduction,
) error {
	if len(production) == 0 {
		return nil
	}

	models := make([]ReturningProduction, 0, len(production))
	for _, p := range production {
		if p == nil {
			continue
		}
		models = append(models, ReturningProduction{
			Season:              p.Season,
			Team:                p.Team,
			Conference:          p.Conference,
			TotalPPA:            p.Total_PPA,
			TotalPassingPPA:     p.TotalPassing_PPA,
			TotalReceivingPPA:   p.TotalReceiving_PPA,
			TotalRushingPPA:     p.TotalRushing_PPA,
			PercentPPA:          p.Percent_PPA,
			PercentPassingPPA:   p.PercentPassing_PPA,
			PercentReceivingPPA: p.PercentReceiving_PPA,
			PercentRushingPPA:   p.PercentRushing_PPA,
			Usage:               p.Usage,
			PassingUsage:        p.PassingUsage,
			ReceivingUsage:      p.ReceivingUsage,
			RushingUsage:        p.RushingUsage,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertPlayerTransfers inserts player transfers.
func (db *Database) InsertPlayerTransfers(
	ctx context.Context,
	transfers []*cfbd.PlayerTransfer,
) error {
	if len(transfers) == 0 {
		return nil
	}

	models := make([]PlayerTransfer, 0, len(transfers))
	for _, t := range transfers {
		if t == nil {
			continue
		}

		var transferDate *time.Time
		if t.TransferDate != nil {
			ts := t.TransferDate.AsTime()
			transferDate = &ts
		}

		models = append(models, PlayerTransfer{
			Season:       t.Season,
			FirstName:    t.FirstName,
			LastName:     t.LastName,
			Position:     t.Position,
			Origin:       t.Origin,
			Destination:  t.Destination,
			TransferDate: transferDate,
			Rating:       t.Rating,
			Stars:        t.Stars,
			Eligibility:  t.Eligibility,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertPlayerStats inserts season player stats.
func (db *Database) InsertPlayerStats(
	ctx context.Context,
	stats []*cfbd.PlayerStat,
) error {
	if len(stats) == 0 {
		return nil
	}

	models := make([]PlayerStat, 0, len(stats))
	for _, s := range stats {
		if s == nil {
			continue
		}

		models = append(models, PlayerStat{
			Season:     s.Season,
			PlayerID:   s.PlayerId, // protobuf field
			Player:     s.Player,
			Team:       s.Team,
			Conference: s.Conference,
			Category:   s.Category,
			StatType:   s.StatType,
			Stat:       s.Stat,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertTeamStats inserts season team stats.
func (db *Database) InsertTeamStats(
	ctx context.Context,
	stats []*cfbd.TeamStat,
) error {
	if len(stats) == 0 {
		return nil
	}

	models := make([]TeamStat, 0, len(stats))
	for _, s := range stats {
		if s == nil {
			continue
		}

		val, err := json.Marshal(s.StatValue)
		if err != nil {
			slog.Error("failed to marshal team stat value", "err", err)
			continue
		}

		models = append(models, TeamStat{
			Season:     s.Season,
			Team:       s.Team,
			Conference: s.Conference,
			StatName:   s.StatName,
			StatValue:  datatypes.JSON(val),
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertRankings inserts poll rankings.
func (db *Database) InsertRankings(
	ctx context.Context,
	weeks []*cfbd.PollWeek,
) error {
	if len(weeks) == 0 {
		return nil
	}

	models := make([]PollWeek, 0, len(weeks))
	for _, pw := range weeks {
		if pw == nil {
			continue
		}

		polls := make([]Poll, 0, len(pw.Polls))
		for _, p := range pw.Polls {
			if p == nil {
				continue
			}

			ranks := make([]PollRank, 0, len(p.Ranks))
			for _, r := range p.Ranks {
				if r == nil {
					continue
				}
				ranks = append(ranks, PollRank{
					Rank:            r.Rank,
					School:          r.School,
					Conference:      r.Conference,
					FirstPlaceVotes: r.FirstPlaceVotes,
					Points:          r.Points,
				})
			}

			polls = append(polls, Poll{
				Poll:  p.Poll,
				Ranks: ranks,
			})
		}

		models = append(models, PollWeek{
			Season:     pw.Season,
			SeasonType: pw.SeasonType,
			Week:       pw.Week,
			Polls:      polls,
		})
	}

	// Reduced batch size for complex associations
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 20).Error
}

// InsertRecruits inserts recruiting data.
func (db *Database) InsertRecruits(
	ctx context.Context,
	recruits []*cfbd.Recruit,
) error {
	if len(recruits) == 0 {
		return nil
	}

	models := make([]Recruit, 0, len(recruits))
	for _, r := range recruits {
		if r == nil {
			continue
		}

		var hometownInfo *RecruitHometownInfo
		if r.HometownInfo != nil {
			hometownInfo = &RecruitHometownInfo{
				FIPSCode:  r.HometownInfo.FipsCode,
				Longitude: r.HometownInfo.Longitude,
				Latitude:  r.HometownInfo.Latitude,
			}
		}

		models = append(models, Recruit{
			ID:            r.Id, // string ID from API
			AthleteID:     r.AthleteId,
			RecruitType:   r.RecruitType,
			Year:          r.Year,
			Ranking:       r.Ranking,
			Name:          r.Name,
			School:        r.School,
			CommittedTo:   r.CommittedTo,
			Position:      r.Position,
			Height:        r.Height,
			Weight:        r.Weight,
			Stars:         r.Stars,
			Rating:        r.Rating,
			City:          r.City,
			StateProvince: r.StateProvince,
			Country:       r.Country,
			HometownInfo:  hometownInfo,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertTeamRecruitingRankings inserts team recruiting rankings.
func (db *Database) InsertTeamRecruitingRankings(
	ctx context.Context,
	rankings []*cfbd.TeamRecruitingRanking,
) error {
	if len(rankings) == 0 {
		return nil
	}

	models := make([]TeamRecruitingRanking, 0, len(rankings))
	for _, r := range rankings {
		if r == nil {
			continue
		}
		models = append(models, TeamRecruitingRanking{
			Year:   r.Year,
			Rank:   r.Rank,
			Team:   r.Team,
			Points: r.Points,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertDraftPicks inserts NFL draft picks.
func (db *Database) InsertDraftPicks(
	ctx context.Context,
	picks []*cfbd.DraftPick,
) error {
	if len(picks) == 0 {
		return nil
	}

	models := make([]DraftPick, 0, len(picks))
	for _, p := range picks {
		if p == nil {
			continue
		}

		var hometownInfo *DraftPickHometownInfo
		if p.HometownInfo != nil {
			hometownInfo = &DraftPickHometownInfo{
				CountyFIPS: p.HometownInfo.CountyFips,
				Longitude:  p.HometownInfo.Longitude,
				Latitude:   p.HometownInfo.Latitude,
				Country:    p.HometownInfo.Country,
				State:      p.HometownInfo.State,
				City:       p.HometownInfo.City,
			}
		}

		models = append(models, DraftPick{
			CollegeAthleteID:        p.CollegeAthleteId,
			NflAthleteID:            p.NflAthleteId,
			CollegeID:               p.CollegeId,
			CollegeTeam:             p.CollegeTeam,
			CollegeConference:       p.CollegeConference,
			NflTeamID:               p.NflTeamId,
			NflTeam:                 p.NflTeam,
			Year:                    p.Year,
			Overall:                 p.Overall,
			Round:                   p.Round,
			Pick:                    p.Pick,
			Name:                    p.Name,
			Position:                p.Position,
			Height:                  p.Height,
			Weight:                  p.Weight,
			PreDraftRanking:         p.PreDraftRanking,
			PreDraftPositionRanking: p.PreDraftPositionRanking,
			PreDraftGrade:           p.PreDraftGrade,
			HometownInfo:            hometownInfo,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertGameTeamStats inserts game team stats.
func (db *Database) InsertGameTeamStats(ctx context.Context, stats []*cfbd.GameTeamStats) error {
	if len(stats) == 0 {
		return nil
	}

	models := make([]GameTeamStats, 0, len(stats))
	for _, s := range stats {
		if s == nil {
			continue
		}

		teams := make([]GameTeamStatsTeam, 0, len(s.Teams))
		for _, t := range s.Teams {
			if t == nil {
				continue
			}

			subStats := make([]GameTeamStatsTeamStat, 0, len(t.Stats))
			for _, st := range t.Stats {
				if st == nil {
					continue
				}
				subStats = append(subStats, GameTeamStatsTeamStat{
					Category: st.Category,
					Stat:     st.Stat,
				})
			}

			teams = append(teams, GameTeamStatsTeam{
				TeamID:     t.TeamId,
				Team:       t.Team,
				Conference: t.Conference,
				HomeAway:   t.HomeAway,
				Points:     t.Points,
				Stats:      subStats,
			})
		}

		models = append(models, GameTeamStats{
			ID:    s.Id,
			Teams: teams,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 50).Error
}

// InsertGamePlayerStats inserts game player stats.
func (db *Database) InsertGamePlayerStats(
	ctx context.Context,
	stats []*cfbd.GamePlayerStats,
) error {
	if len(stats) == 0 {
		return nil
	}

	models := make([]GamePlayerStats, 0, len(stats))
	for _, s := range stats {
		if s == nil {
			continue
		}

		teams := make([]GamePlayerStatsTeam, 0, len(s.Teams))
		for _, t := range s.Teams {
			if t == nil {
				continue
			}

			cats := make([]GamePlayerStatCategories, 0, len(t.Categories))
			for _, c := range t.Categories {
				if c == nil {
					continue
				}

				types := make([]GamePlayerStatTypes, 0, len(c.Types))
				for _, typ := range c.Types {
					if typ == nil {
						continue
					}

					athletes := make([]GamePlayerStatPlayer, 0, len(typ.Athletes))
					for _, a := range typ.Athletes {
						if a == nil {
							continue
						}
						athletes = append(athletes, GamePlayerStatPlayer{
							PlayerID: a.Id,
							Name:     a.Name,
							Stat:     a.Stat,
						})
					}

					types = append(types, GamePlayerStatTypes{
						Name:     typ.Name,
						Athletes: athletes,
					})
				}

				cats = append(cats, GamePlayerStatCategories{
					Name:  c.Name,
					Types: types,
				})
			}

			teams = append(teams, GamePlayerStatsTeam{
				Team:       t.Team,
				Conference: t.Conference,
				HomeAway:   t.HomeAway,
				Points:     t.Points,
				Categories: cats,
			})
		}

		models = append(models, GamePlayerStats{
			ID:    s.Id,
			Teams: teams,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 20).Error // Smaller batch for very deep nesting
}

// GetGameIDs returns a slice of game IDs for a given season.
func (db *Database) GetGameIDs(ctx context.Context, year int) ([]int32, error) {
	var ids []int32
	err := db.WithContext(ctx).Model(&Game{}).
		Where("season = ?", year).
		Pluck("id", &ids).Error
	return ids, err
}

// InsertPlayWinProbability inserts play win probabilities.
func (db *Database) InsertPlayWinProbability(
	ctx context.Context,
	plays []*cfbd.PlayWinProbability,
) error {
	if len(plays) == 0 {
		return nil
	}

	models := make([]PlayWinProbability, 0, len(plays))
	for _, p := range plays {
		if p == nil {
			continue
		}
		models = append(models, PlayWinProbability{
			GameID:             p.GameId,
			PlayID:             p.PlayId,
			PlayText:           p.PlayText,
			HomeID:             p.HomeId,
			Home:               p.Home,
			AwayID:             p.AwayId,
			Away:               p.Away,
			Spread:             p.Spread,
			HomeBall:           p.HomeBall,
			HomeScore:          p.HomeScore,
			AwayScore:          p.AwayScore,
			YardLine:           p.YardLine,
			Down:               p.Down,
			Distance:           p.Distance,
			HomeWinProbability: p.HomeWinProbability,
			PlayNumber:         p.PlayNumber,
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

// InsertAdvancedBoxScores inserts advanced box scores.
func (db *Database) InsertAdvancedBoxScores(
	ctx context.Context,
	scores map[int32]*cfbd.AdvancedBoxScore,
) error {
	if len(scores) == 0 {
		return nil
	}

	models := make([]AdvancedBoxScore, 0, len(scores))
	for gameID, val := range scores {
		if val == nil {
			continue
		}

		payload, err := json.Marshal(val)
		if err != nil {
			slog.Error(
				"failed to marshal advanced box score",
				"err", err,
				"game_id", gameID,
			)
			continue
		}

		models = append(models, AdvancedBoxScore{
			GameID:  gameID,
			Payload: datatypes.JSON(payload),
		})
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(models, 100).Error
}

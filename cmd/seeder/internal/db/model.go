package db

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// NOTE:
// - These are Postgres-oriented GORM v2 models for the CFBD proto messages.
// - Table names are schema-qualified as cfbd.<table>.
// - For repeated/nested structures, child tables are used where reasonable; for dynamic/unknown
//   shapes (google.protobuf.Struct / google.protobuf.Value), jsonb columns are used.
// - Many “stat blob” style endpoints are also persisted as jsonb to keep schema stable.
//
// You can AutoMigrate these in dependency order (venues, teams, conferences, games, etc.).

// ============================================================
// Shared / helper (embeddable) structs
// ============================================================

// EpaSplit is embedded in a few metrics tables.
type EpaSplit struct {
	Rushing float64 `gorm:"column:rushing;not null"`
	Passing float64 `gorm:"column:passing;not null"`
	Total   float64 `gorm:"column:total;not null"`
}

// SuccessRateSplit is embedded in a few metrics tables.
type SuccessRateSplit struct {
	PassingDowns  float64 `gorm:"column:passing_downs;not null"`
	StandardDowns float64 `gorm:"column:standard_downs;not null"`
	Total         float64 `gorm:"column:total;not null"`
}

// RushingYardsSplit is embedded in a few metrics tables.
type RushingYardsSplit struct {
	HighlightYards   float64 `gorm:"column:highlight_yards;not null"`
	OpenFieldYards   float64 `gorm:"column:open_field_yards;not null"`
	SecondLevelYards float64 `gorm:"column:second_level_yards;not null"`
	LineYards        float64 `gorm:"column:line_yards;not null"`
}

// ClockInt32 is used by plays/drives.
type ClockInt32 struct {
	Seconds *int32 `gorm:"column:seconds"`
	Minutes *int32 `gorm:"column:minutes"`
}

// ClockDouble is used by play stats.
type ClockDouble struct {
	Seconds *float64 `gorm:"column:seconds"`
	Minutes *float64 `gorm:"column:minutes"`
}

// StatValue stores google.protobuf.Value as jsonb
type StatValue struct {
	Value datatypes.JSON `gorm:"column:value;type:jsonb"`
}

func (StatValue) TableName() string { return "stat_values" }

// ============================================================
// Reference / dimensions
// ============================================================

type Venue struct {
	ID               int32    `gorm:"primaryKey;column:id"`
	Name             string   `gorm:"column:name;not null"`
	City             string   `gorm:"column:city"`
	State            string   `gorm:"column:state"`
	Zip              string   `gorm:"column:zip"`
	CountryCode      string   `gorm:"column:country_code"`
	Timezone         string   `gorm:"column:timezone"`
	Latitude         *float64 `gorm:"column:latitude"`
	Longitude        *float64 `gorm:"column:longitude"`
	Elevation        string   `gorm:"column:elevation"`
	Capacity         *int32   `gorm:"column:capacity"`
	ConstructionYear *int32   `gorm:"column:construction_year"`
	Grass            *bool    `gorm:"column:grass"`
	Dome             *bool    `gorm:"column:dome"`
}

func (Venue) TableName() string { return "venues" }

type Conference struct {
	ID             int32  `gorm:"primaryKey;column:id"`
	Name           string `gorm:"column:name;not null"`
	ShortName      string `gorm:"column:short_name"`
	Abbreviation   string `gorm:"column:abbreviation"`
	Classification string `gorm:"column:classification"`
}

func (Conference) TableName() string { return "conferences" }

type Team struct {
	ID             int32          `gorm:"primaryKey;column:id"`
	School         string         `gorm:"column:school;not null"`
	Mascot         string         `gorm:"column:mascot"`
	Abbreviation   string         `gorm:"column:abbreviation"`
	AlternateNames pq.StringArray `gorm:"column:alternate_names;type:text[]"`
	Conference     string         `gorm:"column:conference"`
	Division       string         `gorm:"column:division"`
	Classification string         `gorm:"column:classification"`
	Color          string         `gorm:"column:color"`
	AlternateColor string         `gorm:"column:alternate_color"`
	Logos          pq.StringArray `gorm:"column:logos;type:text[]"`
	Twitter        string         `gorm:"column:twitter"`
	VenueID        *int32         `gorm:"column:venue_id;index"`

	Venue *Venue `gorm:"foreignKey:VenueID;references:ID"`
}

func (Team) TableName() string { return "teams" }

// ============================================================
// Games (core spine)
// ============================================================

type Game struct {
	ID                     int32         `gorm:"primaryKey;column:id"`
	Season                 int32         `gorm:"column:season;index;not null"`
	Week                   int32         `gorm:"column:week;index;not null"`
	SeasonType             string        `gorm:"column:season_type;index;not null"`
	StartDate              *time.Time    `gorm:"column:start_date;index"`
	StartTimeTBD           bool          `gorm:"column:start_time_tbd;not null"`
	Completed              bool          `gorm:"column:completed;index;not null"`
	NeutralSite            bool          `gorm:"column:neutral_site;not null"`
	ConferenceGame         bool          `gorm:"column:conference_game;not null"`
	Attendance             *int32        `gorm:"column:attendance"`
	VenueID                *int32        `gorm:"column:venue_id;index"`
	Venue                  string        `gorm:"column:venue"`
	HomeID                 *int32        `gorm:"column:home_id;index"`
	HomeTeam               string        `gorm:"column:home_team"`
	HomeConference         string        `gorm:"column:home_conference"`
	HomeClassification     string        `gorm:"column:home_classification"`
	HomePoints             *int32        `gorm:"column:home_points"`
	HomeLineScores         pq.Int64Array `gorm:"column:home_line_scores;type:int[]"`
	HomePostWinProbability *float64      `gorm:"column:home_postgame_win_probability"`
	HomePregameElo         *int32        `gorm:"column:home_pregame_elo"`
	HomePostgameElo        *int32        `gorm:"column:home_postgame_elo"`

	AwayID                 *int32        `gorm:"column:away_id;index"`
	AwayTeam               string        `gorm:"column:away_team"`
	AwayConference         string        `gorm:"column:away_conference"`
	AwayClassification     string        `gorm:"column:away_classification"`
	AwayPoints             *int32        `gorm:"column:away_points"`
	AwayLineScores         pq.Int64Array `gorm:"column:away_line_scores;type:int[]"`
	AwayPostWinProbability *float64      `gorm:"column:away_postgame_win_probability"`
	AwayPregameElo         *int32        `gorm:"column:away_pregame_elo"`
	AwayPostgameElo        *int32        `gorm:"column:away_postgame_elo"`

	ExcitementIndex *float64 `gorm:"column:excitement_index"`
	Highlights      string   `gorm:"column:highlights"`
	Notes           string   `gorm:"column:notes"`

	VenueRef *Venue `gorm:"foreignKey:VenueID;references:ID"`
	HomeRef  *Team  `gorm:"foreignKey:HomeID;references:ID"`
	AwayRef  *Team  `gorm:"foreignKey:AwayID;references:ID"`
}

func (Game) TableName() string { return "games" }

// ============================================================
// Matchups
// ============================================================

type Matchup struct {
	MatchupID int64  `gorm:"primaryKey;column:matchup_id"`
	Team1     string `gorm:"column:team1;not null"`
	Team2     string `gorm:"column:team2;not null"`
	StartYear *int   `gorm:"column:start_year"`
	EndYear   *int   `gorm:"column:end_year"`
	Team1Wins int    `gorm:"column:team1_wins;not null"`
	Team2Wins int    `gorm:"column:team2_wins;not null"`
	Ties      int    `gorm:"column:ties;not null"`

	Games []MatchupGame `gorm:"foreignKey:MatchupID;references:MatchupID"`
}

func (Matchup) TableName() string { return "matchups" }

type MatchupGame struct {
	ID          int64  `gorm:"primaryKey;column:id"`
	MatchupID   int64  `gorm:"column:matchup_id;index;not null"`
	Season      int32  `gorm:"column:season;not null"`
	Week        int32  `gorm:"column:week;not null"`
	SeasonType  string `gorm:"column:season_type;not null"`
	Date        string `gorm:"column:date"`
	NeutralSite bool   `gorm:"column:neutral_site;not null"`
	Venue       string `gorm:"column:venue"`
	HomeTeam    string `gorm:"column:home_team;not null"`
	HomeScore   *int32 `gorm:"column:home_score"`
	AwayTeam    string `gorm:"column:away_team;not null"`
	AwayScore   *int32 `gorm:"column:away_score"`
	Winner      string `gorm:"column:winner"`
}

func (MatchupGame) TableName() string { return "matchup_games" }

// ============================================================
// Teams endpoints
// ============================================================

type TeamATS struct {
	Year           int32    `gorm:"primaryKey;column:year"`
	TeamID         int32    `gorm:"primaryKey;column:team_id"`
	Team           string   `gorm:"column:team;not null"`
	Conference     string   `gorm:"column:conference"`
	Games          *int32   `gorm:"column:games"`
	AtsWins        int32    `gorm:"column:ats_wins;not null"`
	AtsLosses      int32    `gorm:"column:ats_losses;not null"`
	AtsPushes      int32    `gorm:"column:ats_pushes;not null"`
	AvgCoverMargin *float64 `gorm:"column:avg_cover_margin"`
}

func (TeamATS) TableName() string { return "team_ats" }

type RosterPlayer struct {
	ID             string         `gorm:"primaryKey;column:id"`
	FirstName      string         `gorm:"column:first_name;not null"`
	LastName       string         `gorm:"column:last_name;not null"`
	Team           string         `gorm:"column:team;index;not null"`
	Height         *float64       `gorm:"column:height"`
	Weight         *int32         `gorm:"column:weight"`
	Jersey         *int32         `gorm:"column:jersey"`
	Position       string         `gorm:"column:position"`
	HomeCity       string         `gorm:"column:home_city"`
	HomeState      string         `gorm:"column:home_state"`
	HomeCountry    string         `gorm:"column:home_country"`
	HomeLatitude   *float64       `gorm:"column:home_latitude"`
	HomeLongitude  *float64       `gorm:"column:home_longitude"`
	HomeCountyFIPS string         `gorm:"column:home_county_fips"`
	RecruitIDs     pq.StringArray `gorm:"column:recruit_ids;type:text[]"`
}

func (RosterPlayer) TableName() string { return "roster_players" }

type TeamTalent struct {
	Year   int32   `gorm:"primaryKey;column:year"`
	Team   string  `gorm:"primaryKey;column:team"`
	Talent float64 `gorm:"column:talent;not null"`
}

func (TeamTalent) TableName() string { return "team_talent" }

// ============================================================
// /records
// ============================================================

type TeamRecord struct {
	Games  int32 `gorm:"column:games;not null"`
	Wins   int32 `gorm:"column:wins;not null"`
	Losses int32 `gorm:"column:losses;not null"`
	Ties   int32 `gorm:"column:ties;not null"`
}

// TeamRecords uses embedded TeamRecord with prefixes for each split.
type TeamRecords struct {
	Year           int32    `gorm:"primaryKey;column:year"`
	Team           string   `gorm:"primaryKey;column:team"`
	TeamID         *int32   `gorm:"column:team_id"`
	Classification string   `gorm:"column:classification"`
	Conference     string   `gorm:"column:conference"`
	Division       string   `gorm:"column:division"`
	ExpectedWins   *float64 `gorm:"column:expected_wins"`

	TotalGames  int32 `gorm:"column:total_games;not null"`
	TotalWins   int32 `gorm:"column:total_wins;not null"`
	TotalLosses int32 `gorm:"column:total_losses;not null"`
	TotalTies   int32 `gorm:"column:total_ties;not null"`

	ConferenceGamesGames  int32 `gorm:"column:conference_games_games;not null"`
	ConferenceGamesWins   int32 `gorm:"column:conference_games_wins;not null"`
	ConferenceGamesLosses int32 `gorm:"column:conference_games_losses;not null"`
	ConferenceGamesTies   int32 `gorm:"column:conference_games_ties;not null"`

	HomeGamesGames  int32 `gorm:"column:home_games_games;not null"`
	HomeGamesWins   int32 `gorm:"column:home_games_wins;not null"`
	HomeGamesLosses int32 `gorm:"column:home_games_losses;not null"`
	HomeGamesTies   int32 `gorm:"column:home_games_ties;not null"`

	AwayGamesGames  int32 `gorm:"column:away_games_games;not null"`
	AwayGamesWins   int32 `gorm:"column:away_games_wins;not null"`
	AwayGamesLosses int32 `gorm:"column:away_games_losses;not null"`
	AwayGamesTies   int32 `gorm:"column:away_games_ties;not null"`

	NeutralSiteGamesGames  int32 `gorm:"column:neutral_site_games_games;not null"`
	NeutralSiteGamesWins   int32 `gorm:"column:neutral_site_games_wins;not null"`
	NeutralSiteGamesLosses int32 `gorm:"column:neutral_site_games_losses;not null"`
	NeutralSiteGamesTies   int32 `gorm:"column:neutral_site_games_ties;not null"`

	RegularSeasonGames  int32 `gorm:"column:regular_season_games;not null"`
	RegularSeasonWins   int32 `gorm:"column:regular_season_wins;not null"`
	RegularSeasonLosses int32 `gorm:"column:regular_season_losses;not null"`
	RegularSeasonTies   int32 `gorm:"column:regular_season_ties;not null"`

	PostseasonGames  int32 `gorm:"column:postseason_games;not null"`
	PostseasonWins   int32 `gorm:"column:postseason_wins;not null"`
	PostseasonLosses int32 `gorm:"column:postseason_losses;not null"`
	PostseasonTies   int32 `gorm:"column:postseason_ties;not null"`
}

func (TeamRecords) TableName() string { return "team_records" }

// ============================================================
// /calendar
// ============================================================

type CalendarWeek struct {
	Season         int32      `gorm:"primaryKey;column:season"`
	Week           int32      `gorm:"primaryKey;column:week"`
	SeasonType     string     `gorm:"primaryKey;column:season_type"`
	StartDate      *time.Time `gorm:"column:start_date"`
	EndDate        *time.Time `gorm:"column:end_date"`
	FirstGameStart *time.Time `gorm:"column:first_game_start"`
	LastGameStart  *time.Time `gorm:"column:last_game_start"`
}

func (CalendarWeek) TableName() string { return "calendar_weeks" }

// ============================================================
// /scoreboard (Struct-heavy, stored as jsonb payload)
// ============================================================

type Scoreboard struct {
	ID             int32          `gorm:"primaryKey;column:id"`
	StartDate      *time.Time     `gorm:"column:start_date"`
	StartTimeTBD   bool           `gorm:"column:start_time_tbd;not null"`
	TV             string         `gorm:"column:tv"`
	NeutralSite    bool           `gorm:"column:neutral_site;not null"`
	ConferenceGame bool           `gorm:"column:conference_game;not null"`
	Status         string         `gorm:"column:status"`
	Period         *int32         `gorm:"column:period"`
	Clock          string         `gorm:"column:clock"`
	Situation      string         `gorm:"column:situation"`
	Possession     string         `gorm:"column:possession"`
	LastPlay       string         `gorm:"column:last_play"`
	Venue          datatypes.JSON `gorm:"column:venue;type:jsonb"`
	HomeTeam       datatypes.JSON `gorm:"column:home_team;type:jsonb"`
	AwayTeam       datatypes.JSON `gorm:"column:away_team;type:jsonb"`
	Weather        datatypes.JSON `gorm:"column:weather;type:jsonb"`
	Betting        datatypes.JSON `gorm:"column:betting;type:jsonb"`
}

func (Scoreboard) TableName() string { return "scoreboard" }

// ============================================================
// Drives & Plays
// ============================================================

type Drive struct {
	ID                string `gorm:"primaryKey;column:id"`
	GameID            int32  `gorm:"column:game_id;index;not null"`
	Offense           string `gorm:"column:offense"`
	OffenseConference string `gorm:"column:offense_conference"`
	Defense           string `gorm:"column:defense"`
	DefenseConference string `gorm:"column:defense_conference"`
	DriveNumber       *int32 `gorm:"column:drive_number;index"`
	Scoring           bool   `gorm:"column:scoring;not null"`
	StartPeriod       int32  `gorm:"column:start_period;not null"`
	StartYardline     int32  `gorm:"column:start_yardline;not null"`
	StartYardsToGoal  int32  `gorm:"column:start_yards_to_goal;not null"`
	StartTimeMinutes  *int32 `gorm:"column:start_time_minutes"`
	StartTimeSeconds  *int32 `gorm:"column:start_time_seconds"`
	EndPeriod         int32  `gorm:"column:end_period;not null"`
	EndYardline       int32  `gorm:"column:end_yardline;not null"`
	EndYardsToGoal    int32  `gorm:"column:end_yards_to_goal;not null"`
	EndTimeMinutes    *int32 `gorm:"column:end_time_minutes"`
	EndTimeSeconds    *int32 `gorm:"column:end_time_seconds"`
	ElapsedMinutes    *int32 `gorm:"column:elapsed_minutes"`
	ElapsedSeconds    *int32 `gorm:"column:elapsed_seconds"`
	Plays             int32  `gorm:"column:plays;not null"`
	Yards             int32  `gorm:"column:yards;not null"`
	DriveResult       string `gorm:"column:drive_result"`
	IsHomeOffense     bool   `gorm:"column:is_home_offense;not null"`
	StartOffenseScore int32  `gorm:"column:start_offense_score;not null"`
	StartDefenseScore int32  `gorm:"column:start_defense_score;not null"`
	EndOffenseScore   int32  `gorm:"column:end_offense_score;not null"`
	EndDefenseScore   int32  `gorm:"column:end_defense_score;not null"`
}

func (Drive) TableName() string { return "drives" }

type Play struct {
	ID                string   `gorm:"primaryKey;column:id"`
	DriveID           string   `gorm:"column:drive_id;index"`
	GameID            int32    `gorm:"column:game_id;index;not null"`
	DriveNumber       *int32   `gorm:"column:drive_number"`
	PlayNumber        *int32   `gorm:"column:play_number;index"`
	Offense           string   `gorm:"column:offense;index"`
	OffenseConference string   `gorm:"column:offense_conference"`
	OffenseScore      int32    `gorm:"column:offense_score;not null"`
	Defense           string   `gorm:"column:defense;index"`
	Home              string   `gorm:"column:home"`
	Away              string   `gorm:"column:away"`
	DefenseConference string   `gorm:"column:defense_conference"`
	DefenseScore      int32    `gorm:"column:defense_score;not null"`
	Period            int32    `gorm:"column:period;index;not null"`
	ClockMinutes      *int32   `gorm:"column:clock_minutes"`
	ClockSeconds      *int32   `gorm:"column:clock_seconds"`
	OffenseTimeouts   *int32   `gorm:"column:offense_timeouts"`
	DefenseTimeouts   *int32   `gorm:"column:defense_timeouts"`
	Yardline          int32    `gorm:"column:yardline;not null"`
	YardsToGoal       int32    `gorm:"column:yards_to_goal;not null"`
	Down              int32    `gorm:"column:down;index;not null"`
	Distance          int32    `gorm:"column:distance;not null"`
	YardsGained       int32    `gorm:"column:yards_gained;not null"`
	Scoring           bool     `gorm:"column:scoring;index;not null"`
	PlayType          string   `gorm:"column:play_type;index"`
	PlayText          string   `gorm:"column:play_text"`
	PPA               *float64 `gorm:"column:ppa"`
	Wallclock         string   `gorm:"column:wallclock"`
}

func (Play) TableName() string { return "plays" }

type PlayType struct {
	ID           int32  `gorm:"primaryKey;column:id"`
	Text         string `gorm:"column:text;not null"`
	Abbreviation string `gorm:"column:abbreviation"`
}

func (PlayType) TableName() string { return "play_types" }

// ============================================================
// /plays/stats
// ============================================================

type PlayStat struct {
	ID            int64    `gorm:"primaryKey;column:id"`
	GameID        float64  `gorm:"column:game_id;index"`
	Season        float64  `gorm:"column:season;index"`
	Week          float64  `gorm:"column:week;index"`
	Team          string   `gorm:"column:team;index"`
	Conference    string   `gorm:"column:conference"`
	Opponent      string   `gorm:"column:opponent"`
	TeamScore     float64  `gorm:"column:team_score"`
	OpponentScore float64  `gorm:"column:opponent_score"`
	DriveID       string   `gorm:"column:drive_id;index"`
	PlayID        string   `gorm:"column:play_id;index"`
	Period        float64  `gorm:"column:period"`
	ClockMinutes  *float64 `gorm:"column:clock_minutes"`
	ClockSeconds  *float64 `gorm:"column:clock_seconds"`
	YardsToGoal   float64  `gorm:"column:yards_to_goal"`
	Down          float64  `gorm:"column:down"`
	Distance      float64  `gorm:"column:distance"`
	AthleteID     string   `gorm:"column:athlete_id;index"`
	AthleteName   string   `gorm:"column:athlete_name"`
	StatType      string   `gorm:"column:stat_type;index"`
	Stat          float64  `gorm:"column:stat"`
}

func (PlayStat) TableName() string { return "play_stats" }

type PlayStatType struct {
	ID   int32  `gorm:"primaryKey;column:id"`
	Name string `gorm:"column:name;not null"`
}

func (PlayStatType) TableName() string { return "play_stat_types" }

// ============================================================
// Players
// ============================================================

type PlayerSearchResult struct {
	ID                 string   `gorm:"primaryKey;column:id"`
	Team               string   `gorm:"column:team;index"`
	Name               string   `gorm:"column:name;not null"`
	FirstName          string   `gorm:"column:first_name"`
	LastName           string   `gorm:"column:last_name"`
	Weight             *int32   `gorm:"column:weight"`
	Height             *float64 `gorm:"column:height"`
	Jersey             *int32   `gorm:"column:jersey"`
	Position           string   `gorm:"column:position;index"`
	Hometown           string   `gorm:"column:hometown"`
	TeamColor          string   `gorm:"column:team_color"`
	TeamColorSecondary string   `gorm:"column:team_color_secondary"`
}

func (PlayerSearchResult) TableName() string { return "player_search_results" }

type PlayerPPAChartItem struct {
	ID         int64   `gorm:"primaryKey;column:id"`
	PlayerID   string  `gorm:"column:player_id;index"`
	PlayNumber int32   `gorm:"column:play_number;not null"`
	AvgPPA     float64 `gorm:"column:avg_ppa;not null"`
}

func (PlayerPPAChartItem) TableName() string { return "player_ppa_chart_items" }

type PlayerUsageSplits struct {
	ID            int64    `gorm:"primaryKey;column:id"`
	PassingDowns  *float64 `gorm:"column:passing_downs"`
	StandardDowns *float64 `gorm:"column:standard_downs"`
	ThirdDown     *float64 `gorm:"column:third_down"`
	SecondDown    *float64 `gorm:"column:second_down"`
	FirstDown     *float64 `gorm:"column:first_down"`
	Rush          *float64 `gorm:"column:rush"`
	Pass          *float64 `gorm:"column:pass"`
	Overall       *float64 `gorm:"column:overall"`
}

func (PlayerUsageSplits) TableName() string { return "player_usage_splits" }

type PlayerUsage struct {
	Season     int32  `gorm:"primaryKey;column:season"`
	ID         string `gorm:"primaryKey;column:id"`
	Name       string `gorm:"column:name;not null"`
	Position   string `gorm:"column:position;index"`
	Team       string `gorm:"column:team;index"`
	Conference string `gorm:"column:conference"`

	UsageID *int64             `gorm:"column:usage_id;index"`
	Usage   *PlayerUsageSplits `gorm:"foreignKey:UsageID;references:ID"`
}

func (PlayerUsage) TableName() string { return "player_usage" }

type ReturningProduction struct {
	Season     int32  `gorm:"primaryKey;column:season"`
	Team       string `gorm:"primaryKey;column:team"`
	Conference string `gorm:"column:conference"`

	TotalPPA            float64 `gorm:"column:total_ppa;not null"`
	TotalPassingPPA     float64 `gorm:"column:total_passing_ppa;not null"`
	TotalReceivingPPA   float64 `gorm:"column:total_receiving_ppa;not null"`
	TotalRushingPPA     float64 `gorm:"column:total_rushing_ppa;not null"`
	PercentPPA          float64 `gorm:"column:percent_ppa;not null"`
	PercentPassingPPA   float64 `gorm:"column:percent_passing_ppa;not null"`
	PercentReceivingPPA float64 `gorm:"column:percent_receiving_ppa;not null"`
	PercentRushingPPA   float64 `gorm:"column:percent_rushing_ppa;not null"`
	Usage               float64 `gorm:"column:usage;not null"`
	PassingUsage        float64 `gorm:"column:passing_usage;not null"`
	ReceivingUsage      float64 `gorm:"column:receiving_usage;not null"`
	RushingUsage        float64 `gorm:"column:rushing_usage;not null"`
}

func (ReturningProduction) TableName() string { return "returning_production" }

type PlayerTransfer struct {
	Season       int32      `gorm:"primaryKey;column:season"`
	FirstName    string     `gorm:"primaryKey;column:first_name"`
	LastName     string     `gorm:"primaryKey;column:last_name"`
	Position     string     `gorm:"column:position"`
	Origin       string     `gorm:"column:origin"`
	Destination  string     `gorm:"column:destination"`
	TransferDate *time.Time `gorm:"column:transfer_date"`
	Rating       *float64   `gorm:"column:rating"`
	Stars        *int32     `gorm:"column:stars"`
	Eligibility  string     `gorm:"column:eligibility"`
}

func (PlayerTransfer) TableName() string { return "player_transfers" }

// ============================================================
// /stats/player/season and /stats/season
// ============================================================

type PlayerStat struct {
	ID         int64  `gorm:"primaryKey;column:id"`
	Season     int32  `gorm:"column:season;index;not null"`
	PlayerID   string `gorm:"column:player_id;index;not null"`
	Player     string `gorm:"column:player"`
	Position   string `gorm:"column:position;index"`
	Team       string `gorm:"column:team;index"`
	Conference string `gorm:"column:conference"`
	Category   string `gorm:"column:category;index"`
	StatType   string `gorm:"column:stat_type;index"`
	Stat       string `gorm:"column:stat"`
}

func (PlayerStat) TableName() string { return "player_stats" }

type TeamStat struct {
	ID         int64          `gorm:"primaryKey;column:id"`
	Season     int32          `gorm:"column:season;index;not null"`
	Team       string         `gorm:"column:team;index;not null"`
	Conference string         `gorm:"column:conference"`
	StatName   string         `gorm:"column:stat_name;index;not null"`
	StatValue  datatypes.JSON `gorm:"column:stat_value;type:jsonb"`
}

func (TeamStat) TableName() string { return "team_stats" }

// ============================================================
// Advanced season/game stats
// Persisted as jsonb for stability; full sub-structures also modeled.
// ============================================================

type AdvancedRateMetrics struct {
	ID            int64    `gorm:"primaryKey;column:id"`
	Explosiveness *float64 `gorm:"column:explosiveness"`
	SuccessRate   *float64 `gorm:"column:success_rate"`
	TotalPPA      *float64 `gorm:"column:total_ppa"`
	PPA           *float64 `gorm:"column:ppa"`
	Rate          *float64 `gorm:"column:rate"`
}

func (AdvancedRateMetrics) TableName() string { return "advanced_rate_metrics" }

type AdvancedHavoc struct {
	ID         int64    `gorm:"primaryKey;column:id"`
	DB         *float64 `gorm:"column:db"`
	FrontSeven *float64 `gorm:"column:front_seven"`
	Total      *float64 `gorm:"column:total"`
}

func (AdvancedHavoc) TableName() string { return "advanced_havoc" }

type AdvancedFieldPosition struct {
	ID                     int64    `gorm:"primaryKey;column:id"`
	AveragePredictedPoints *float64 `gorm:"column:average_predicted_points"`
	AverageStart           *float64 `gorm:"column:average_start"`
}

func (AdvancedFieldPosition) TableName() string { return "advanced_field_position" }

type AdvancedSeasonStatSide struct {
	ID      int64          `gorm:"primaryKey;column:id"`
	Payload datatypes.JSON `gorm:"column:payload;type:jsonb"`
}

func (AdvancedSeasonStatSide) TableName() string { return "advanced_season_stat_sides" }

type AdvancedSeasonStat struct {
	Season     int32  `gorm:"primaryKey;column:season"`
	Team       string `gorm:"primaryKey;column:team"`
	Conference string `gorm:"column:conference"`

	OffenseSideID *int64                  `gorm:"column:offense_side_id;index"`
	DefenseSideID *int64                  `gorm:"column:defense_side_id;index"`
	Offense       *AdvancedSeasonStatSide `gorm:"foreignKey:OffenseSideID;references:ID"`
	Defense       *AdvancedSeasonStatSide `gorm:"foreignKey:DefenseSideID;references:ID"`
}

func (AdvancedSeasonStat) TableName() string { return "advanced_season_stats" }

type AdvancedGameStatSide struct {
	ID      int64          `gorm:"primaryKey;column:id"`
	Payload datatypes.JSON `gorm:"column:payload;type:jsonb"`
}

func (AdvancedGameStatSide) TableName() string { return "advanced_game_stat_sides" }

type AdvancedGameStat struct {
	GameID     int32  `gorm:"primaryKey;column:game_id"`
	Season     int32  `gorm:"column:season;index"`
	SeasonType string `gorm:"column:season_type;index"`
	Week       int32  `gorm:"column:week;index"`
	Team       string `gorm:"column:team;index"`
	Opponent   string `gorm:"column:opponent;index"`

	OffenseSideID *int64                `gorm:"column:offense_side_id;index"`
	DefenseSideID *int64                `gorm:"column:defense_side_id;index"`
	Offense       *AdvancedGameStatSide `gorm:"foreignKey:OffenseSideID;references:ID"`
	Defense       *AdvancedGameStatSide `gorm:"foreignKey:DefenseSideID;references:ID"`
}

func (AdvancedGameStat) TableName() string { return "advanced_game_stats" }

type GameHavocStatSide struct {
	ID                    int64   `gorm:"primaryKey;column:id"`
	DBHavocRate           float64 `gorm:"column:db_havoc_rate;not null"`
	FrontSevenHavocRate   float64 `gorm:"column:front_seven_havoc_rate;not null"`
	HavocRate             float64 `gorm:"column:havoc_rate;not null"`
	DBHavocEvents         float64 `gorm:"column:db_havoc_events;not null"`
	FrontSevenHavocEvents float64 `gorm:"column:front_seven_havoc_events;not null"`
	TotalHavocEvents      float64 `gorm:"column:total_havoc_events;not null"`
	TotalPlays            float64 `gorm:"column:total_plays;not null"`
}

func (GameHavocStatSide) TableName() string { return "game_havoc_stat_sides" }

type GameHavocStats struct {
	GameID             int32  `gorm:"primaryKey;column:game_id"`
	Season             int32  `gorm:"column:season;index"`
	SeasonType         string `gorm:"column:season_type;index"`
	Week               int32  `gorm:"column:week;index"`
	Team               string `gorm:"column:team;index"`
	Conference         string `gorm:"column:conference"`
	Opponent           string `gorm:"column:opponent;index"`
	OpponentConference string `gorm:"column:opponent_conference"`

	OffenseID *int64             `gorm:"column:offense_id;index"`
	DefenseID *int64             `gorm:"column:defense_id;index"`
	Offense   *GameHavocStatSide `gorm:"foreignKey:OffenseID;references:ID"`
	Defense   *GameHavocStatSide `gorm:"foreignKey:DefenseID;references:ID"`
}

func (GameHavocStats) TableName() string { return "game_havoc_stats" }

// ============================================================
// Recruiting
// ============================================================

type RecruitHometownInfo struct {
	ID        int64    `gorm:"primaryKey;column:id"`
	FIPSCode  string   `gorm:"column:fips_code"`
	Longitude *float64 `gorm:"column:longitude"`
	Latitude  *float64 `gorm:"column:latitude"`
}

func (RecruitHometownInfo) TableName() string { return "recruit_hometown_info" }

type Recruit struct {
	ID            string   `gorm:"primaryKey;column:id"`
	AthleteID     string   `gorm:"column:athlete_id;index"`
	RecruitType   string   `gorm:"column:recruit_type;index"`
	Year          int32    `gorm:"column:year;index;not null"`
	Ranking       *int32   `gorm:"column:ranking"`
	Name          string   `gorm:"column:name;not null"`
	School        string   `gorm:"column:school"`
	CommittedTo   string   `gorm:"column:committed_to;index"`
	Position      string   `gorm:"column:position;index"`
	Height        *float64 `gorm:"column:height"`
	Weight        *int32   `gorm:"column:weight"`
	Stars         int32    `gorm:"column:stars;not null"`
	Rating        float64  `gorm:"column:rating;not null"`
	City          string   `gorm:"column:city"`
	StateProvince string   `gorm:"column:state_province"`
	Country       string   `gorm:"column:country"`

	HometownInfoID *int64               `gorm:"column:hometown_info_id;index"`
	HometownInfo   *RecruitHometownInfo `gorm:"foreignKey:HometownInfoID;references:ID"`
}

func (Recruit) TableName() string { return "recruits" }

type TeamRecruitingRanking struct {
	Year   int32   `gorm:"primaryKey;column:year"`
	Team   string  `gorm:"primaryKey;column:team"`
	Rank   int32   `gorm:"column:rank;not null"`
	Points float64 `gorm:"column:points;not null"`
}

func (TeamRecruitingRanking) TableName() string { return "team_recruiting_rankings" }

type AggregatedTeamRecruiting struct {
	Team          string  `gorm:"primaryKey;column:team"`
	Conference    string  `gorm:"primaryKey;column:conference"`
	PositionGroup string  `gorm:"primaryKey;column:position_group"`
	AverageRating float64 `gorm:"column:average_rating;not null"`
	TotalRating   float64 `gorm:"column:total_rating;not null"`
	Commits       int32   `gorm:"column:commits;not null"`
	AverageStars  float64 `gorm:"column:average_stars;not null"`
}

func (AggregatedTeamRecruiting) TableName() string { return "aggregated_team_recruiting" }

// ============================================================
// Ratings: SP / SRS / Elo / FPI
// Stored largely as jsonb payloads, with primary keys on (year, team|conference).
// ============================================================

type TeamSP struct {
	Year       int32          `gorm:"primaryKey;column:year"`
	Team       string         `gorm:"primaryKey;column:team"`
	Conference string         `gorm:"column:conference"`
	Payload    datatypes.JSON `gorm:"column:payload;type:jsonb"`
}

func (TeamSP) TableName() string { return "team_sp" }

type ConferenceSP struct {
	Year       int32          `gorm:"primaryKey;column:year"`
	Conference string         `gorm:"primaryKey;column:conference"`
	Payload    datatypes.JSON `gorm:"column:payload;type:jsonb"`
}

func (ConferenceSP) TableName() string { return "conference_sp" }

type TeamSRS struct {
	Year       int32   `gorm:"primaryKey;column:year"`
	Team       string  `gorm:"primaryKey;column:team"`
	Conference string  `gorm:"column:conference"`
	Division   string  `gorm:"column:division"`
	Rating     float64 `gorm:"column:rating;not null"`
	Ranking    *int32  `gorm:"column:ranking"`
}

func (TeamSRS) TableName() string { return "team_srs" }

type TeamElo struct {
	Year       int32  `gorm:"primaryKey;column:year"`
	Team       string `gorm:"primaryKey;column:team"`
	Conference string `gorm:"column:conference"`
	Elo        *int32 `gorm:"column:elo"`
}

func (TeamElo) TableName() string { return "team_elo" }

type TeamFPI struct {
	Year       int32          `gorm:"primaryKey;column:year"`
	Team       string         `gorm:"primaryKey;column:team"`
	Conference string         `gorm:"column:conference"`
	Payload    datatypes.JSON `gorm:"column:payload;type:jsonb"`
}

func (TeamFPI) TableName() string { return "team_fpi" }

// ============================================================
// Polls / rankings
// ============================================================

type PollWeek struct {
	ID         int64  `gorm:"primaryKey;column:id"`
	Season     int32  `gorm:"column:season;index;not null"`
	SeasonType string `gorm:"column:season_type;index;not null"`
	Week       int32  `gorm:"column:week;index;not null"`

	Polls []Poll `gorm:"foreignKey:PollWeekID;references:ID"`
}

func (PollWeek) TableName() string { return "poll_weeks" }

type Poll struct {
	ID         int64  `gorm:"primaryKey;column:id"`
	PollWeekID int64  `gorm:"column:poll_week_id;index;not null"`
	Poll       string `gorm:"column:poll;not null"`

	Ranks []PollRank `gorm:"foreignKey:PollID;references:ID"`
}

func (Poll) TableName() string { return "polls" }

type PollRank struct {
	ID              int64  `gorm:"primaryKey;column:id"`
	PollID          int64  `gorm:"column:poll_id;index;not null"`
	Rank            *int32 `gorm:"column:rank"`
	TeamID          *int32 `gorm:"column:team_id"`
	School          string `gorm:"column:school;not null"`
	Conference      string `gorm:"column:conference"`
	FirstPlaceVotes *int32 `gorm:"column:first_place_votes"`
	Points          *int32 `gorm:"column:points"`
}

func (PollRank) TableName() string { return "poll_ranks" }

// ============================================================
// Betting / lines
// ============================================================

type BettingGame struct {
	ID                 int32      `gorm:"primaryKey;column:id"`
	Season             int32      `gorm:"column:season;index;not null"`
	SeasonType         string     `gorm:"column:season_type;index;not null"`
	Week               int32      `gorm:"column:week;index;not null"`
	StartDate          *time.Time `gorm:"column:start_date"`
	HomeTeamID         int32      `gorm:"column:home_team_id;index"`
	HomeTeam           string     `gorm:"column:home_team"`
	HomeConference     string     `gorm:"column:home_conference"`
	HomeClassification string     `gorm:"column:home_classification"`
	HomeScore          *int32     `gorm:"column:home_score"`
	AwayTeamID         int32      `gorm:"column:away_team_id;index"`
	AwayTeam           string     `gorm:"column:away_team"`
	AwayConference     string     `gorm:"column:away_conference"`
	AwayClassification string     `gorm:"column:away_classification"`
	AwayScore          *int32     `gorm:"column:away_score"`

	Lines []GameLine `gorm:"foreignKey:GameID;references:ID"`
}

func (BettingGame) TableName() string { return "betting_games" }

type GameLine struct {
	GameID          int32    `gorm:"primaryKey;column:game_id"`
	Provider        string   `gorm:"primaryKey;column:provider"`
	Spread          *float64 `gorm:"column:spread"`
	FormattedSpread string   `gorm:"column:formatted_spread"`
	SpreadOpen      *float64 `gorm:"column:spread_open"`
	OverUnder       *float64 `gorm:"column:over_under"`
	OverUnderOpen   *float64 `gorm:"column:over_under_open"`
	HomeMoneyline   *float64 `gorm:"column:home_moneyline"`
	AwayMoneyline   *float64 `gorm:"column:away_moneyline"`
}

func (GameLine) TableName() string { return "game_lines" }

// ============================================================
// Media & Weather
// ============================================================

type GameMedia struct {
	ID             int32      `gorm:"primaryKey;column:id"`
	Season         int32      `gorm:"column:season;index"`
	Week           int32      `gorm:"column:week;index"`
	SeasonType     string     `gorm:"column:season_type;index"`
	StartTime      *time.Time `gorm:"column:start_time"`
	IsStartTimeTBD bool       `gorm:"column:is_start_time_tbd;not null"`
	HomeTeam       string     `gorm:"column:home_team"`
	HomeConference string     `gorm:"column:home_conference"`
	AwayTeam       string     `gorm:"column:away_team"`
	AwayConference string     `gorm:"column:away_conference"`
	MediaType      string     `gorm:"column:media_type"`
	Outlet         string     `gorm:"column:outlet"`
}

func (GameMedia) TableName() string { return "game_media" }

type GameWeather struct {
	ID                   int32      `gorm:"primaryKey;column:id"`
	Season               int32      `gorm:"column:season;index"`
	Week                 int32      `gorm:"column:week;index"`
	SeasonType           string     `gorm:"column:season_type;index"`
	StartTime            *time.Time `gorm:"column:start_time"`
	GameIndoors          bool       `gorm:"column:game_indoors;not null"`
	HomeTeam             string     `gorm:"column:home_team"`
	HomeConference       string     `gorm:"column:home_conference"`
	AwayTeam             string     `gorm:"column:away_team"`
	AwayConference       string     `gorm:"column:away_conference"`
	VenueID              *int32     `gorm:"column:venue_id;index"`
	Venue                string     `gorm:"column:venue"`
	Temperature          *float64   `gorm:"column:temperature"`
	DewPoint             *float64   `gorm:"column:dew_point"`
	Humidity             *float64   `gorm:"column:humidity"`
	Precipitation        *float64   `gorm:"column:precipitation"`
	Snowfall             *float64   `gorm:"column:snowfall"`
	WindDirection        *float64   `gorm:"column:wind_direction"`
	WindSpeed            *float64   `gorm:"column:wind_speed"`
	Pressure             *float64   `gorm:"column:pressure"`
	WeatherConditionCode *float64   `gorm:"column:weather_condition_code"`
	WeatherCondition     string     `gorm:"column:weather_condition"`
}

func (GameWeather) TableName() string { return "game_weather" }

// ============================================================
// Game team stats (box score)
//
// Modeled as nested tables:
// GameTeamStats(id) -> many teams -> many stats
// ============================================================

type GameTeamStats struct {
	ID int32 `gorm:"primaryKey;column:id"`

	Teams []GameTeamStatsTeam `gorm:"foreignKey:GameID;references:ID"`
}

func (GameTeamStats) TableName() string { return "game_team_stats" }

type GameTeamStatsTeam struct {
	ID         int64  `gorm:"primaryKey;column:id"`
	GameID     int32  `gorm:"column:game_id;index;not null"`
	TeamID     int32  `gorm:"column:team_id;index;not null"`
	Team       string `gorm:"column:team;not null"`
	Conference string `gorm:"column:conference"`
	HomeAway   string `gorm:"column:home_away"`
	Points     *int32 `gorm:"column:points"`

	Stats []GameTeamStatsTeamStat `gorm:"foreignKey:TeamRowID;references:ID"`
}

func (GameTeamStatsTeam) TableName() string { return "game_team_stats_teams" }

type GameTeamStatsTeamStat struct {
	ID        int64  `gorm:"primaryKey;column:id"`
	TeamRowID int64  `gorm:"column:team_row_id;index;not null"`
	Category  string `gorm:"column:category;index;not null"`
	Stat      string `gorm:"column:stat;not null"`
}

func (GameTeamStatsTeamStat) TableName() string { return "game_team_stats_team_stats" }

// ============================================================
// Game player stats (very nested)
//
// GamePlayerStats(id) -> teams -> categories -> types -> athletes
// ============================================================

type GamePlayerStats struct {
	ID int32 `gorm:"primaryKey;column:id"`

	Teams []GamePlayerStatsTeam `gorm:"foreignKey:GameID;references:ID"`
}

func (GamePlayerStats) TableName() string { return "game_player_stats" }

type GamePlayerStatsTeam struct {
	ID         int64  `gorm:"primaryKey;column:id"`
	GameID     int32  `gorm:"column:game_id;index;not null"`
	Team       string `gorm:"column:team;index;not null"`
	Conference string `gorm:"column:conference"`
	HomeAway   string `gorm:"column:home_away"`
	Points     *int32 `gorm:"column:points"`

	Categories []GamePlayerStatCategories `gorm:"foreignKey:TeamRowID;references:ID"`
}

func (GamePlayerStatsTeam) TableName() string { return "game_player_stats_teams" }

type GamePlayerStatCategories struct {
	ID        int64  `gorm:"primaryKey;column:id"`
	TeamRowID int64  `gorm:"column:team_row_id;index;not null"`
	Name      string `gorm:"column:name;index;not null"`

	Types []GamePlayerStatTypes `gorm:"foreignKey:CategoryRowID;references:ID"`
}

func (GamePlayerStatCategories) TableName() string { return "game_player_stat_categories" }

type GamePlayerStatTypes struct {
	ID            int64  `gorm:"primaryKey;column:id"`
	CategoryRowID int64  `gorm:"column:category_row_id;index;not null"`
	Name          string `gorm:"column:name;index;not null"`

	Athletes []GamePlayerStatPlayer `gorm:"foreignKey:TypeRowID;references:ID"`
}

func (GamePlayerStatTypes) TableName() string { return "game_player_stat_types" }

type GamePlayerStatPlayer struct {
	ID        int64  `gorm:"primaryKey;column:id"`
	TypeRowID int64  `gorm:"column:type_row_id;index;not null"`
	PlayerID  string `gorm:"column:player_id;index;not null"`
	Name      string `gorm:"column:name;not null"`
	Stat      string `gorm:"column:stat;not null"`
}

func (GamePlayerStatPlayer) TableName() string { return "game_player_stat_players" }

// ============================================================
// Live game (/live/plays) nested entities
// ============================================================

type LiveGame struct {
	ID          int32  `gorm:"primaryKey;column:id"`
	Status      string `gorm:"column:status"`
	Period      *int32 `gorm:"column:period"`
	Clock       string `gorm:"column:clock"`
	Possession  string `gorm:"column:possession"`
	Down        *int32 `gorm:"column:down"`
	Distance    *int32 `gorm:"column:distance"`
	YardsToGoal *int32 `gorm:"column:yards_to_goal"`

	Teams  []LiveGameTeam  `gorm:"foreignKey:LiveGameID;references:ID"`
	Drives []LiveGameDrive `gorm:"foreignKey:LiveGameID;references:ID"`
}

func (LiveGame) TableName() string { return "live_games" }

type LiveGameTeam struct {
	ID                      int64         `gorm:"primaryKey;column:id"`
	LiveGameID              int32         `gorm:"column:live_game_id;index;not null"`
	TeamID                  int32         `gorm:"column:team_id;index;not null"`
	Team                    string        `gorm:"column:team;not null"`
	HomeAway                string        `gorm:"column:home_away"`
	LineScores              pq.Int64Array `gorm:"column:line_scores;type:int[]"`
	Points                  int32         `gorm:"column:points;not null"`
	Drives                  int32         `gorm:"column:drives;not null"`
	ScoringOpportunities    int32         `gorm:"column:scoring_opportunities;not null"`
	PointsPerOpportunity    float64       `gorm:"column:points_per_opportunity;not null"`
	AverageStartYardLine    *float64      `gorm:"column:average_start_yard_line"`
	Plays                   int32         `gorm:"column:plays;not null"`
	LineYards               float64       `gorm:"column:line_yards;not null"`
	LineYardsPerRush        float64       `gorm:"column:line_yards_per_rush;not null"`
	SecondLevelYards        float64       `gorm:"column:second_level_yards;not null"`
	SecondLevelYardsPerRush float64       `gorm:"column:second_level_yards_per_rush;not null"`
	OpenFieldYards          float64       `gorm:"column:open_field_yards;not null"`
	OpenFieldYardsPerRush   float64       `gorm:"column:open_field_yards_per_rush;not null"`
	EpaPerPlay              float64       `gorm:"column:epa_per_play;not null"`
	TotalEpa                float64       `gorm:"column:total_epa;not null"`
	PassingEpa              float64       `gorm:"column:passing_epa;not null"`
	EpaPerPass              float64       `gorm:"column:epa_per_pass;not null"`
	RushingEpa              float64       `gorm:"column:rushing_epa;not null"`
	EpaPerRush              float64       `gorm:"column:epa_per_rush;not null"`
	SuccessRate             float64       `gorm:"column:success_rate;not null"`
	StandardDownSuccessRate float64       `gorm:"column:standard_down_success_rate;not null"`
	PassingDownSuccessRate  float64       `gorm:"column:passing_down_success_rate;not null"`
	Explosiveness           float64       `gorm:"column:explosiveness;not null"`
	DeserveToWin            *float64      `gorm:"column:deserve_to_win"`
}

func (LiveGameTeam) TableName() string { return "live_game_teams" }

type LiveGameDrive struct {
	ID                 string `gorm:"primaryKey;column:id"`
	LiveGameID         int32  `gorm:"column:live_game_id;index;not null"`
	OffenseID          int32  `gorm:"column:offense_id"`
	Offense            string `gorm:"column:offense"`
	DefenseID          int32  `gorm:"column:defense_id"`
	Defense            string `gorm:"column:defense"`
	PlayCount          int32  `gorm:"column:play_count;not null"`
	Yards              int32  `gorm:"column:yards;not null"`
	StartPeriod        int32  `gorm:"column:start_period;not null"`
	StartClock         string `gorm:"column:start_clock"`
	StartYardsToGoal   int32  `gorm:"column:start_yards_to_goal;not null"`
	EndPeriod          *int32 `gorm:"column:end_period"`
	EndClock           string `gorm:"column:end_clock"`
	EndYardsToGoal     *int32 `gorm:"column:end_yards_to_goal"`
	Duration           string `gorm:"column:duration"`
	ScoringOpportunity bool   `gorm:"column:scoring_opportunity;not null"`
	Result             string `gorm:"column:result"`
	PointsGained       int32  `gorm:"column:points_gained;not null"`

	Plays []LiveGamePlay `gorm:"foreignKey:DriveID;references:ID"`
}

func (LiveGameDrive) TableName() string { return "live_game_drives" }

type LiveGamePlay struct {
	ID          string     `gorm:"primaryKey;column:id"`
	DriveID     string     `gorm:"column:drive_id;index;not null"`
	HomeScore   int32      `gorm:"column:home_score;not null"`
	AwayScore   int32      `gorm:"column:away_score;not null"`
	Period      int32      `gorm:"column:period;not null"`
	Clock       string     `gorm:"column:clock"`
	WallClock   *time.Time `gorm:"column:wall_clock"`
	TeamID      int32      `gorm:"column:team_id"`
	Team        string     `gorm:"column:team"`
	Down        int32      `gorm:"column:down"`
	Distance    int32      `gorm:"column:distance"`
	YardsToGoal int32      `gorm:"column:yards_to_goal"`
	YardsGained int32      `gorm:"column:yards_gained"`
	PlayTypeID  int32      `gorm:"column:play_type_id"`
	PlayType    string     `gorm:"column:play_type"`
	Epa         *float64   `gorm:"column:epa"`
	GarbageTime bool       `gorm:"column:garbage_time;not null"`
	Success     bool       `gorm:"column:success;not null"`
	RushPass    string     `gorm:"column:rush_pass"`
	DownType    string     `gorm:"column:down_type"`
	PlayText    string     `gorm:"column:play_text"`
}

func (LiveGamePlay) TableName() string { return "live_game_plays" }

// ============================================================
// PPA predicted points & PPA endpoints
// ============================================================

type PredictedPointsValue struct {
	Down            int32   `gorm:"primaryKey;column:down"`
	Distance        int32   `gorm:"primaryKey;column:distance"`
	YardLine        int32   `gorm:"primaryKey;column:yard_line"`
	PredictedPoints float64 `gorm:"column:predicted_points;not null"`
}

func (PredictedPointsValue) TableName() string { return "predicted_points_values" }

type TeamSeasonPredictedPointsAdded struct {
	Season     int32          `gorm:"primaryKey;column:season"`
	Conference string         `gorm:"primaryKey;column:conference"`
	Team       string         `gorm:"primaryKey;column:team"`
	Offense    datatypes.JSON `gorm:"column:offense;type:jsonb"`
	Defense    datatypes.JSON `gorm:"column:defense;type:jsonb"`
}

func (TeamSeasonPredictedPointsAdded) TableName() string { return "team_season_ppa" }

type TeamGamePredictedPointsAdded struct {
	GameID     int32          `gorm:"primaryKey;column:game_id"`
	Season     int32          `gorm:"column:season;index"`
	Week       int32          `gorm:"column:week;index"`
	SeasonType string         `gorm:"column:season_type;index"`
	Team       string         `gorm:"column:team;index"`
	Conference string         `gorm:"column:conference"`
	Opponent   string         `gorm:"column:opponent;index"`
	Offense    datatypes.JSON `gorm:"column:offense;type:jsonb"`
	Defense    datatypes.JSON `gorm:"column:defense;type:jsonb"`
}

func (TeamGamePredictedPointsAdded) TableName() string { return "team_game_ppa" }

type PlayerGamePredictedPointsAdded struct {
	Season     int32          `gorm:"primaryKey;column:season"`
	Week       int32          `gorm:"primaryKey;column:week"`
	SeasonType string         `gorm:"primaryKey;column:season_type"`
	PlayerID   string         `gorm:"primaryKey;column:player_id"`
	Name       string         `gorm:"column:name"`
	Position   string         `gorm:"column:position"`
	Team       string         `gorm:"column:team;index"`
	Opponent   string         `gorm:"column:opponent;index"`
	AveragePPA datatypes.JSON `gorm:"column:average_ppa;type:jsonb"`
}

func (PlayerGamePredictedPointsAdded) TableName() string { return "player_game_ppa" }

type PlayerSeasonPredictedPointsAdded struct {
	Season     int32          `gorm:"primaryKey;column:season"`
	PlayerID   string         `gorm:"primaryKey;column:player_id"`
	Name       string         `gorm:"column:name"`
	Position   string         `gorm:"column:position"`
	Team       string         `gorm:"column:team;index"`
	Conference string         `gorm:"column:conference"`
	AveragePPA datatypes.JSON `gorm:"column:average_ppa;type:jsonb"`
	TotalPPA   datatypes.JSON `gorm:"column:total_ppa;type:jsonb"`
}

func (PlayerSeasonPredictedPointsAdded) TableName() string { return "player_season_ppa" }

// ============================================================
// Win probability
// ============================================================

type PlayWinProbability struct {
	GameID             int32   `gorm:"primaryKey;column:game_id"`
	PlayID             string  `gorm:"primaryKey;column:play_id"`
	PlayText           string  `gorm:"column:play_text"`
	HomeID             int32   `gorm:"column:home_id"`
	Home               string  `gorm:"column:home"`
	AwayID             int32   `gorm:"column:away_id"`
	Away               string  `gorm:"column:away"`
	Spread             float64 `gorm:"column:spread"`
	HomeBall           bool    `gorm:"column:home_ball;not null"`
	HomeScore          int32   `gorm:"column:home_score;not null"`
	AwayScore          int32   `gorm:"column:away_score;not null"`
	YardLine           int32   `gorm:"column:yard_line;not null"`
	Down               int32   `gorm:"column:down;not null"`
	Distance           int32   `gorm:"column:distance;not null"`
	HomeWinProbability float64 `gorm:"column:home_win_probability;not null"`
	PlayNumber         int32   `gorm:"column:play_number;not null"`
}

func (PlayWinProbability) TableName() string { return "play_win_probability" }

type PregameWinProbability struct {
	GameID             int32   `gorm:"primaryKey;column:game_id"`
	Season             int32   `gorm:"column:season;index"`
	SeasonType         string  `gorm:"column:season_type;index"`
	Week               int32   `gorm:"column:week;index"`
	HomeTeam           string  `gorm:"column:home_team"`
	AwayTeam           string  `gorm:"column:away_team"`
	Spread             float64 `gorm:"column:spread"`
	HomeWinProbability float64 `gorm:"column:home_win_probability"`
}

func (PregameWinProbability) TableName() string { return "pregame_win_probability" }

type FieldGoalEP struct {
	YardsToGoal    int32   `gorm:"primaryKey;column:yards_to_goal"`
	Distance       int32   `gorm:"primaryKey;column:distance"`
	ExpectedPoints float64 `gorm:"column:expected_points;not null"`
}

func (FieldGoalEP) TableName() string { return "field_goal_ep" }

// ============================================================
// Advanced box score (nested & wide) stored as jsonb payload
// ============================================================

type AdvancedBoxScore struct {
	GameID  int32          `gorm:"primaryKey;column:game_id"`
	Payload datatypes.JSON `gorm:"column:payload;type:jsonb"`
}

func (AdvancedBoxScore) TableName() string { return "advanced_box_scores" }

// ============================================================
// Draft
// ============================================================

type DraftTeam struct {
	ID          int64  `gorm:"primaryKey;column:id"`
	Location    string `gorm:"column:location"`
	Nickname    string `gorm:"column:nickname"`
	DisplayName string `gorm:"column:display_name"`
	Logo        string `gorm:"column:logo"`
}

func (DraftTeam) TableName() string { return "draft_teams" }

type DraftPosition struct {
	ID           int64  `gorm:"primaryKey;column:id"`
	Name         string `gorm:"column:name"`
	Abbreviation string `gorm:"column:abbreviation"`
}

func (DraftPosition) TableName() string { return "draft_positions" }

type DraftPickHometownInfo struct {
	ID         int64  `gorm:"primaryKey;column:id"`
	CountyFIPS string `gorm:"column:county_fips"`
	Longitude  string `gorm:"column:longitude"`
	Latitude   string `gorm:"column:latitude"`
	Country    string `gorm:"column:country"`
	State      string `gorm:"column:state"`
	City       string `gorm:"column:city"`
}

func (DraftPickHometownInfo) TableName() string { return "draft_pick_hometown_info" }

type DraftPick struct {
	ID                      int64    `gorm:"primaryKey;column:id"`
	CollegeAthleteID        *int32   `gorm:"column:college_athlete_id"`
	NflAthleteID            *int32   `gorm:"column:nfl_athlete_id"`
	CollegeID               int32    `gorm:"column:college_id;index;not null"`
	CollegeTeam             string   `gorm:"column:college_team"`
	CollegeConference       string   `gorm:"column:college_conference"`
	NflTeamID               int32    `gorm:"column:nfl_team_id;index;not null"`
	NflTeam                 string   `gorm:"column:nfl_team"`
	Year                    int32    `gorm:"column:year;index;not null"`
	Overall                 int32    `gorm:"column:overall;not null"`
	Round                   int32    `gorm:"column:round;not null"`
	Pick                    int32    `gorm:"column:pick;not null"`
	Name                    string   `gorm:"column:name;not null"`
	Position                string   `gorm:"column:position"`
	Height                  *float64 `gorm:"column:height"`
	Weight                  *int32   `gorm:"column:weight"`
	PreDraftRanking         *int32   `gorm:"column:pre_draft_ranking"`
	PreDraftPositionRanking *int32   `gorm:"column:pre_draft_position_ranking"`
	PreDraftGrade           *int32   `gorm:"column:pre_draft_grade"`

	HometownInfoID *int64                 `gorm:"column:hometown_info_id;index"`
	HometownInfo   *DraftPickHometownInfo `gorm:"foreignKey:HometownInfoID;references:ID"`
}

func (DraftPick) TableName() string { return "draft_picks" }

// ============================================================
// Coaches
// ============================================================

type Coach struct {
	ID        int64      `gorm:"primaryKey;column:id"`
	FirstName string     `gorm:"column:first_name;not null"`
	LastName  string     `gorm:"column:last_name;not null"`
	HireDate  *time.Time `gorm:"column:hire_date"`

	Seasons []CoachSeason `gorm:"foreignKey:CoachID;references:ID"`
}

func (Coach) TableName() string { return "coaches" }

type CoachSeason struct {
	ID             int64    `gorm:"primaryKey;column:id"`
	CoachID        int64    `gorm:"column:coach_id;index;not null"`
	School         string   `gorm:"column:school;index;not null"`
	Year           int32    `gorm:"column:year;index;not null"`
	Games          int32    `gorm:"column:games;not null"`
	Wins           int32    `gorm:"column:wins;not null"`
	Losses         int32    `gorm:"column:losses;not null"`
	Ties           int32    `gorm:"column:ties;not null"`
	PreseasonRank  *int32   `gorm:"column:preseason_rank"`
	PostseasonRank *int32   `gorm:"column:postseason_rank"`
	SRS            *float64 `gorm:"column:srs"`
	SpOverall      *float64 `gorm:"column:sp_overall"`
	SpOffense      *float64 `gorm:"column:sp_offense"`
	SpDefense      *float64 `gorm:"column:sp_defense"`
}

func (CoachSeason) TableName() string { return "coach_seasons" }

// ============================================================
// WEPA
// ============================================================

type AdjustedTeamMetrics struct {
	Year       int32  `gorm:"primaryKey;column:year"`
	TeamID     int32  `gorm:"primaryKey;column:team_id"`
	Team       string `gorm:"column:team;not null"`
	Conference string `gorm:"column:conference"`

	EpaRushing        float64 `gorm:"column:epa_rushing;not null"`
	EpaPassing        float64 `gorm:"column:epa_passing;not null"`
	EpaTotal          float64 `gorm:"column:epa_total;not null"`
	EpaAllowedRushing float64 `gorm:"column:epa_allowed_rushing;not null"`
	EpaAllowedPassing float64 `gorm:"column:epa_allowed_passing;not null"`
	EpaAllowedTotal   float64 `gorm:"column:epa_allowed_total;not null"`

	SuccessRatePassingDowns         float64 `gorm:"column:success_rate_passing_downs;not null"`
	SuccessRateStandardDowns        float64 `gorm:"column:success_rate_standard_downs;not null"`
	SuccessRateTotal                float64 `gorm:"column:success_rate_total;not null"`
	SuccessRateAllowedPassingDowns  float64 `gorm:"column:success_rate_allowed_passing_downs;not null"`
	SuccessRateAllowedStandardDowns float64 `gorm:"column:success_rate_allowed_standard_downs;not null"`
	SuccessRateAllowedTotal         float64 `gorm:"column:success_rate_allowed_total;not null"`

	RushingHighlightYards          float64 `gorm:"column:rushing_highlight_yards;not null"`
	RushingOpenFieldYards          float64 `gorm:"column:rushing_open_field_yards;not null"`
	RushingSecondLevelYards        float64 `gorm:"column:rushing_second_level_yards;not null"`
	RushingLineYards               float64 `gorm:"column:rushing_line_yards;not null"`
	RushingAllowedHighlightYards   float64 `gorm:"column:rushing_allowed_highlight_yards;not null"`
	RushingAllowedOpenFieldYards   float64 `gorm:"column:rushing_allowed_open_field_yards;not null"`
	RushingAllowedSecondLevelYards float64 `gorm:"column:rushing_allowed_second_level_yards;not null"`
	RushingAllowedLineYards        float64 `gorm:"column:rushing_allowed_line_yards;not null"`

	Explosiveness        float64 `gorm:"column:explosiveness;not null"`
	ExplosivenessAllowed float64 `gorm:"column:explosiveness_allowed;not null"`
}

func (AdjustedTeamMetrics) TableName() string { return "adjusted_team_metrics" }

type PlayerWeightedEPA struct {
	Year        int32   `gorm:"primaryKey;column:year"`
	AthleteID   string  `gorm:"primaryKey;column:athlete_id"`
	AthleteName string  `gorm:"column:athlete_name;not null"`
	Position    string  `gorm:"column:position;index"`
	Team        string  `gorm:"column:team;index"`
	Conference  string  `gorm:"column:conference"`
	WEPA        float64 `gorm:"column:wepa;not null"`
	Plays       int32   `gorm:"column:plays;not null"`
}

func (PlayerWeightedEPA) TableName() string { return "player_weighted_epa" }

type KickerPAAR struct {
	Year        int32   `gorm:"primaryKey;column:year"`
	AthleteID   string  `gorm:"primaryKey;column:athlete_id"`
	AthleteName string  `gorm:"column:athlete_name;not null"`
	Team        string  `gorm:"column:team;index"`
	Conference  string  `gorm:"column:conference"`
	PAAR        float64 `gorm:"column:paar;not null"`
	Attempts    int32   `gorm:"column:attempts;not null"`
}

func (KickerPAAR) TableName() string { return "kicker_paar" }

// ============================================================
// Misc endpoints
// ============================================================

type UserInfo struct {
	ID             int64   `gorm:"primaryKey;column:id"`
	PatronLevel    float64 `gorm:"column:patron_level;not null"`
	RemainingCalls float64 `gorm:"column:remaining_calls;not null"`
}

func (UserInfo) TableName() string { return "user_info" }

type Int32List struct {
	ID     int64         `gorm:"primaryKey;column:id"`
	Values pq.Int64Array `gorm:"column:values;type:int[]"`
}

func (Int32List) TableName() string { return "int32_lists" }

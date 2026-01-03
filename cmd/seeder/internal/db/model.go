// models.go
//
// GORM models for schema.sql
// Notes:
// - Schema-qualified tables via TableName() -> "cfbd.<table>".
// - JSONB columns use gorm.io/datatypes.JSON.
// - BIGSERIAL/BIGINT IDs are modeled as int64.
// - Composite PK tables use multiple `primaryKey` tags.
// - Some constraints (like CHECK(kind IN ...)) are best added via db.Exec after AutoMigrate.

package db

import (
   "time"

   "gorm.io/datatypes"
)

// ===========================
// Core reference tables
// ===========================

type Venue struct {
   ID               int      `gorm:"primaryKey;column:id"`
   Name             *string  `gorm:"column:name"`
   City             *string  `gorm:"column:city"`
   State            *string  `gorm:"column:state"`
   Zip              *string  `gorm:"column:zip"`
   CountryCode      *string  `gorm:"column:country_code"`
   Timezone         *string  `gorm:"column:timezone"`
   Latitude         *float64 `gorm:"column:latitude"`
   Longitude        *float64 `gorm:"column:longitude"`
   Elevation        *string  `gorm:"column:elevation"`
   Capacity         *int     `gorm:"column:capacity"`
   ConstructionYear *int     `gorm:"column:construction_year"`
   Grass            *bool    `gorm:"column:grass"`
   Dome             *bool    `gorm:"column:dome"`
}

func (Venue) TableName() string { return "cfbd.venues" }

type Team struct {
   ID             int            `gorm:"primaryKey;column:id"`
   School         string         `gorm:"column:school;not null"`
   Mascot         *string        `gorm:"column:mascot"`
   Abbreviation   *string        `gorm:"column:abbreviation"`
   AlternateNames datatypes.JSON `gorm:"column:alternate_names;type:jsonb"`
   Conference     *string        `gorm:"column:conference"`
   Division       *string        `gorm:"column:division"`
   Classification *string        `gorm:"column:classification"`
   Color          *string        `gorm:"column:color"`
   AlternateColor *string        `gorm:"column:alternate_color"`
   Logos          datatypes.JSON `gorm:"column:logos;type:jsonb"`
   Twitter        string         `gorm:"column:twitter;not null"`

   VenueID *int   `gorm:"column:venue_id"`
   Venue   *Venue `gorm:"foreignKey:VenueID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

func (Team) TableName() string { return "cfbd.teams" }

type Conference struct {
   ID             int     `gorm:"primaryKey;column:id"`
   Name           string  `gorm:"column:name;not null"`
   ShortName      *string `gorm:"column:short_name"`
   Abbreviation   *string `gorm:"column:abbreviation"`
   Classification *string `gorm:"column:classification"`
}

func (Conference) TableName() string { return "cfbd.conferences" }

// ===========================
// Simple top-level metrics
// ===========================

type AdjustedTeamMetrics struct {
   Year       int    `gorm:"primaryKey;column:year"`
   TeamID     int    `gorm:"primaryKey;column:team_id"`
   Team       string `gorm:"column:team;not null"`
   Conference string `gorm:"column:conference;not null"`

   EpaRushing float64 `gorm:"column:epa_rushing;not null"`
   EpaPassing float64 `gorm:"column:epa_passing;not null"`
   EpaTotal   float64 `gorm:"column:epa_total;not null"`

   EpaAllowedRushing float64 `gorm:"column:epa_allowed_rushing;not null"`
   EpaAllowedPassing float64 `gorm:"column:epa_allowed_passing;not null"`
   EpaAllowedTotal   float64 `gorm:"column:epa_allowed_total;not null"`

   SuccessRatePassingDowns  float64 `gorm:"column:success_rate_passing_downs;not null"`
   SuccessRateStandardDowns float64 `gorm:"column:success_rate_standard_downs;not null"`
   SuccessRateTotal         float64 `gorm:"column:success_rate_total;not null"`

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

func (AdjustedTeamMetrics) TableName() string { return "cfbd.adjusted_team_metrics" }

type PlayerWeightedEPA struct {
   Year        int     `gorm:"primaryKey;column:year"`
   AthleteID   string  `gorm:"primaryKey;column:athlete_id"`
   AthleteName string  `gorm:"column:athlete_name;not null"`
   Position    string  `gorm:"column:position;not null"`
   Team        string  `gorm:"column:team;not null"`
   Conference  string  `gorm:"column:conference;not null"`
   WEPA        float64 `gorm:"column:wepa;not null"`
   Plays       int     `gorm:"column:plays;not null"`
}

func (PlayerWeightedEPA) TableName() string { return "cfbd.player_weighted_epa" }

type KickerPAAR struct {
   Year        int     `gorm:"primaryKey;column:year"`
   AthleteID   string  `gorm:"primaryKey;column:athlete_id"`
   AthleteName string  `gorm:"column:athlete_name;not null"`
   Team        string  `gorm:"column:team;not null"`
   Conference  string  `gorm:"column:conference;not null"`
   PAAR        float64 `gorm:"column:paar;not null"`
   Attempts    int     `gorm:"column:attempts;not null"`
}

func (KickerPAAR) TableName() string { return "cfbd.kicker_paar" }

type TeamATS struct {
   Year           int      `gorm:"primaryKey;column:year"`
   TeamID         int      `gorm:"primaryKey;column:team_id"`
   Team           string   `gorm:"column:team;not null"`
   Conference     *string  `gorm:"column:conference"`
   Games          *int     `gorm:"column:games"`
   ATSWins        int      `gorm:"column:ats_wins;not null"`
   ATSLosses      int      `gorm:"column:ats_losses;not null"`
   ATSPushes      int      `gorm:"column:ats_pushes;not null"`
   AvgCoverMargin *float64 `gorm:"column:avg_cover_margin"`
}

func (TeamATS) TableName() string { return "cfbd.team_ats" }

type RosterPlayer struct {
   ID             string         `gorm:"primaryKey;column:id"`
   FirstName      string         `gorm:"column:first_name;not null"`
   LastName       string         `gorm:"column:last_name;not null"`
   Team           string         `gorm:"column:team;not null"`
   Height         *float64       `gorm:"column:height"`
   Weight         *int           `gorm:"column:weight"`
   Jersey         *int           `gorm:"column:jersey"`
   Position       *string        `gorm:"column:position"`
   HomeCity       *string        `gorm:"column:home_city"`
   HomeState      *string        `gorm:"column:home_state"`
   HomeCountry    *string        `gorm:"column:home_country"`
   HomeLatitude   *float64       `gorm:"column:home_latitude"`
   HomeLongitude  *float64       `gorm:"column:home_longitude"`
   HomeCountyFIPS *string        `gorm:"column:home_county_fips"`
   RecruitIDs     datatypes.JSON `gorm:"column:recruit_ids;type:jsonb"`
}

func (RosterPlayer) TableName() string { return "cfbd.roster_players" }

type TeamTalent struct {
   Year   int     `gorm:"primaryKey;column:year"`
   Team   string  `gorm:"primaryKey;column:team"`
   Talent float64 `gorm:"column:talent;not null"`
}

func (TeamTalent) TableName() string { return "cfbd.team_talent" }

type PlayerStat struct {
   Season   int    `gorm:"primaryKey;column:season"`
   PlayerID string `gorm:"primaryKey;column:player_id"`
   Category string `gorm:"primaryKey;column:category"`
   StatType string `gorm:"primaryKey;column:stat_type"`
   Stat     string `gorm:"primaryKey;column:stat"`

   Player     string `gorm:"column:player;not null"`
   Position   string `gorm:"column:position;not null"`
   Team       string `gorm:"column:team;not null"`
   Conference string `gorm:"column:conference;not null"`
}

func (PlayerStat) TableName() string { return "cfbd.player_stats" }

type TeamStat struct {
   Season     int            `gorm:"primaryKey;column:season"`
   Team       string         `gorm:"primaryKey;column:team"`
   StatName   string         `gorm:"primaryKey;column:stat_name"`
   Conference string         `gorm:"column:conference;not null"`
   StatValue  datatypes.JSON `gorm:"column:stat_value;type:jsonb;not null"`
}

func (TeamStat) TableName() string { return "cfbd.team_stats" }

// ===========================
// Matchups
// ===========================

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

func (Matchup) TableName() string { return "cfbd.matchups" }

type MatchupGame struct {
   MatchupGameID int64   `gorm:"primaryKey;column:matchup_game_id"`
   MatchupID     int64   `gorm:"column:matchup_id;not null;index"`
   Season        int     `gorm:"column:season;not null"`
   Week          int     `gorm:"column:week;not null"`
   SeasonType    string  `gorm:"column:season_type;not null"`
   Date          string  `gorm:"column:date;not null"`
   NeutralSite   bool    `gorm:"column:neutral_site;not null"`
   Venue         *string `gorm:"column:venue"`
   HomeTeam      string  `gorm:"column:home_team;not null"`
   HomeScore     *int    `gorm:"column:home_score"`
   AwayTeam      string  `gorm:"column:away_team;not null"`
   AwayScore     *int    `gorm:"column:away_score"`
   Winner        *string `gorm:"column:winner"`

   Matchup Matchup `gorm:"foreignKey:MatchupID;references:MatchupID;constraint:OnDelete:CASCADE"`
}

func (MatchupGame) TableName() string { return "cfbd.matchup_games" }

// ===========================
// SP / SRS / Elo / FPI
// ===========================

type TeamSP struct {
   Year       int     `gorm:"primaryKey;column:year"`
   Team       string  `gorm:"primaryKey;column:team"`
   Conference *string `gorm:"column:conference"`

   Rating  *float64 `gorm:"column:rating"`
   Ranking *int     `gorm:"column:ranking"`

   SecondOrderWins *float64 `gorm:"column:second_order_wins"`
   SOS             *float64 `gorm:"column:sos"`

   Offense      datatypes.JSON `gorm:"column:offense;type:jsonb;not null"`
   Defense      datatypes.JSON `gorm:"column:defense;type:jsonb;not null"`
   SpecialTeams datatypes.JSON `gorm:"column:special_teams;type:jsonb;not null"`
}

func (TeamSP) TableName() string { return "cfbd.team_sp" }

type ConferenceSP struct {
   Year       int    `gorm:"primaryKey;column:year"`
   Conference string `gorm:"primaryKey;column:conference"`

   Rating          float64  `gorm:"column:rating;not null"`
   SecondOrderWins float64  `gorm:"column:second_order_wins;not null"`
   SOS             *float64 `gorm:"column:sos"`

   Offense      datatypes.JSON `gorm:"column:offense;type:jsonb;not null"`
   Defense      datatypes.JSON `gorm:"column:defense;type:jsonb;not null"`
   SpecialTeams datatypes.JSON `gorm:"column:special_teams;type:jsonb;not null"`
}

func (ConferenceSP) TableName() string { return "cfbd.conference_sp" }

type TeamSRS struct {
   Year       int     `gorm:"primaryKey;column:year"`
   Team       string  `gorm:"primaryKey;column:team"`
   Conference *string `gorm:"column:conference"`
   Division   *string `gorm:"column:division"`
   Rating     float64 `gorm:"column:rating;not null"`
   Ranking    *int    `gorm:"column:ranking"`
}

func (TeamSRS) TableName() string { return "cfbd.team_srs" }

type TeamElo struct {
   Year       int     `gorm:"primaryKey;column:year"`
   Team       string  `gorm:"primaryKey;column:team"`
   Conference *string `gorm:"column:conference"`
   Elo        *int    `gorm:"column:elo"`
}

func (TeamElo) TableName() string { return "cfbd.team_elo" }

type TeamFPI struct {
   Year       int      `gorm:"primaryKey;column:year"`
   Team       string   `gorm:"primaryKey;column:team"`
   Conference *string  `gorm:"column:conference"`
   FPI        *float64 `gorm:"column:fpi"`

   ResumeRanks  datatypes.JSON `gorm:"column:resume_ranks;type:jsonb;not null"`
   Efficiencies datatypes.JSON `gorm:"column:efficiencies;type:jsonb;not null"`
}

func (TeamFPI) TableName() string { return "cfbd.team_fpi" }

// ===========================
// Polls
// ===========================

type PollWeek struct {
   PollWeekID int64  `gorm:"primaryKey;column:poll_week_id"`
   Season     int    `gorm:"column:season;not null"`
   SeasonType string `gorm:"column:season_type;not null"`
   Week       int    `gorm:"column:week;not null"`

   Polls []Poll `gorm:"foreignKey:PollWeekID;references:PollWeekID"`
}

func (PollWeek) TableName() string { return "cfbd.poll_weeks" }

type Poll struct {
   PollID     int64  `gorm:"primaryKey;column:poll_id"`
   PollWeekID int64  `gorm:"column:poll_week_id;not null;index"`
   PollName   string `gorm:"column:poll_name;not null"`

   Week  PollWeek   `gorm:"foreignKey:PollWeekID;references:PollWeekID;constraint:OnDelete:CASCADE"`
   Ranks []PollRank `gorm:"foreignKey:PollID;references:PollID"`
}

func (Poll) TableName() string { return "cfbd.polls" }

type PollRank struct {
   PollRankID      int64   `gorm:"primaryKey;column:poll_rank_id"`
   PollID          int64   `gorm:"column:poll_id;not null;index"`
   Rank            *int    `gorm:"column:rank"`
   TeamID          *int    `gorm:"column:team_id"`
   School          string  `gorm:"column:school;not null"`
   Conference      *string `gorm:"column:conference"`
   FirstPlaceVotes *int    `gorm:"column:first_place_votes"`
   Points          *int    `gorm:"column:points"`

   Poll Poll `gorm:"foreignKey:PollID;references:PollID;constraint:OnDelete:CASCADE"`
}

func (PollRank) TableName() string { return "cfbd.poll_ranks" }

// ===========================
// Plays & play stats
// ===========================

type Play struct {
   ID                string   `gorm:"primaryKey;column:id"`
   DriveID           string   `gorm:"column:drive_id;not null;index"`
   GameID            int      `gorm:"column:game_id;not null;index"`
   DriveNumber       *int     `gorm:"column:drive_number"`
   PlayNumber        *int     `gorm:"column:play_number"`
   Offense           string   `gorm:"column:offense;not null"`
   OffenseConference *string  `gorm:"column:offense_conference"`
   OffenseScore      int      `gorm:"column:offense_score;not null"`
   Defense           string   `gorm:"column:defense;not null"`
   Home              string   `gorm:"column:home;not null"`
   Away              string   `gorm:"column:away;not null"`
   DefenseConference *string  `gorm:"column:defense_conference"`
   DefenseScore      int      `gorm:"column:defense_score;not null"`
   Period            int      `gorm:"column:period;not null"`
   ClockSeconds      *int     `gorm:"column:clock_seconds"`
   ClockMinutes      *int     `gorm:"column:clock_minutes"`
   OffenseTimeouts   *int     `gorm:"column:offense_timeouts"`
   DefenseTimeouts   *int     `gorm:"column:defense_timeouts"`
   Yardline          int      `gorm:"column:yardline;not null"`
   YardsToGoal       int      `gorm:"column:yards_to_goal;not null"`
   Down              int      `gorm:"column:down;not null"`
   Distance          int      `gorm:"column:distance;not null"`
   YardsGained       int      `gorm:"column:yards_gained;not null"`
   Scoring           bool     `gorm:"column:scoring;not null"`
   PlayType          string   `gorm:"column:play_type;not null"`
   PlayText          *string  `gorm:"column:play_text"`
   PPA               *float64 `gorm:"column:ppa"`
   Wallclock         *string  `gorm:"column:wallclock"`
}

func (Play) TableName() string { return "cfbd.plays" }

type PlayType struct {
   ID           int     `gorm:"primaryKey;column:id"`
   Text         string  `gorm:"column:text;not null"`
   Abbreviation *string `gorm:"column:abbreviation"`
}

func (PlayType) TableName() string { return "cfbd.play_types" }

type PlayStatType struct {
   ID   int    `gorm:"primaryKey;column:id"`
   Name string `gorm:"column:name;not null"`
}

func (PlayStatType) TableName() string { return "cfbd.play_stat_types" }

type PlayStat struct {
   PlayStatID int64 `gorm:"primaryKey;column:play_stat_id"`

   GameID        float64 `gorm:"column:game_id;not null;index"`
   Season        float64 `gorm:"column:season;not null"`
   Week          float64 `gorm:"column:week;not null"`
   Team          string  `gorm:"column:team;not null"`
   Conference    string  `gorm:"column:conference;not null"`
   Opponent      string  `gorm:"column:opponent;not null"`
   TeamScore     float64 `gorm:"column:team_score;not null"`
   OpponentScore float64 `gorm:"column:opponent_score;not null"`
   DriveID       string  `gorm:"column:drive_id;not null"`
   PlayID        string  `gorm:"column:play_id;not null;index"`
   Period        float64 `gorm:"column:period;not null"`

   ClockSeconds *float64 `gorm:"column:clock_seconds"`
   ClockMinutes *float64 `gorm:"column:clock_minutes"`

   YardsToGoal float64 `gorm:"column:yards_to_goal;not null"`
   Down        float64 `gorm:"column:down;not null"`
   Distance    float64 `gorm:"column:distance;not null"`

   AthleteID   string  `gorm:"column:athlete_id;not null"`
   AthleteName string  `gorm:"column:athlete_name;not null"`
   StatType    string  `gorm:"column:stat_type;not null"`
   Stat        float64 `gorm:"column:stat;not null"`
}

func (PlayStat) TableName() string { return "cfbd.play_stats" }

type PlayerSearchResult struct {
   ID                 string   `gorm:"primaryKey;column:id"`
   Team               string   `gorm:"column:team;not null"`
   Name               string   `gorm:"column:name;not null"`
   FirstName          *string  `gorm:"column:first_name"`
   LastName           *string  `gorm:"column:last_name"`
   Weight             *int     `gorm:"column:weight"`
   Height             *float64 `gorm:"column:height"`
   Jersey             *int     `gorm:"column:jersey"`
   Position           string   `gorm:"column:position;not null"`
   Hometown           string   `gorm:"column:hometown;not null"`
   TeamColor          string   `gorm:"column:team_color;not null"`
   TeamColorSecondary string   `gorm:"column:team_color_secondary;not null"`
}

func (PlayerSearchResult) TableName() string { return "cfbd.player_search_results" }

type PlayerUsage struct {
   Season     int            `gorm:"primaryKey;column:season"`
   ID         string         `gorm:"primaryKey;column:id"`
   Name       string         `gorm:"column:name;not null"`
   Position   string         `gorm:"column:position;not null"`
   Team       string         `gorm:"column:team;not null"`
   Conference string         `gorm:"column:conference;not null"`
   Usage      datatypes.JSON `gorm:"column:usage;type:jsonb;not null"`
}

func (PlayerUsage) TableName() string { return "cfbd.player_usage" }

// ===========================
// Returning production & transfers
// ===========================

type ReturningProduction struct {
   Season     int    `gorm:"primaryKey;column:season"`
   Team       string `gorm:"primaryKey;column:team"`
   Conference string `gorm:"column:conference;not null"`

   TotalPPA          float64 `gorm:"column:total_ppa;not null"`
   TotalPassingPPA   float64 `gorm:"column:total_passing_ppa;not null"`
   TotalReceivingPPA float64 `gorm:"column:total_receiving_ppa;not null"`
   TotalRushingPPA   float64 `gorm:"column:total_rushing_ppa;not null"`

   PercentPPA          float64 `gorm:"column:percent_ppa;not null"`
   PercentPassingPPA   float64 `gorm:"column:percent_passing_ppa;not null"`
   PercentReceivingPPA float64 `gorm:"column:percent_receiving_ppa;not null"`
   PercentRushingPPA   float64 `gorm:"column:percent_rushing_ppa;not null"`

   Usage          float64 `gorm:"column:usage;not null"`
   PassingUsage   float64 `gorm:"column:passing_usage;not null"`
   ReceivingUsage float64 `gorm:"column:receiving_usage;not null"`
   RushingUsage   float64 `gorm:"column:rushing_usage;not null"`
}

func (ReturningProduction) TableName() string { return "cfbd.returning_production" }

type PlayerTransfer struct {
   TransferID   int64     `gorm:"primaryKey;column:transfer_id"`
   Season       int       `gorm:"column:season;not null;index"`
   FirstName    string    `gorm:"column:first_name;not null"`
   LastName     string    `gorm:"column:last_name;not null"`
   Position     string    `gorm:"column:position;not null"`
   Origin       string    `gorm:"column:origin;not null"`
   Destination  *string   `gorm:"column:destination"`
   TransferDate time.Time `gorm:"column:transfer_date;type:timestamptz;not null"`
   Rating       *float64  `gorm:"column:rating"`
   Stars        *int      `gorm:"column:stars"`
   Eligibility  *string   `gorm:"column:eligibility"`
}

func (PlayerTransfer) TableName() string { return "cfbd.player_transfers" }

// ===========================
// Predicted points & PPA added
// ===========================

type PredictedPointsValue struct {
   YardLine        int     `gorm:"primaryKey;column:yard_line"`
   PredictedPoints float64 `gorm:"column:predicted_points;not null"`
}

func (PredictedPointsValue) TableName() string { return "cfbd.predicted_points_values" }

type TeamSeasonPredictedPointsAdded struct {
   Season     int            `gorm:"primaryKey;column:season"`
   Team       string         `gorm:"primaryKey;column:team"`
   Conference string         `gorm:"column:conference;not null"`
   Offense    datatypes.JSON `gorm:"column:offense;type:jsonb;not null"`
   Defense    datatypes.JSON `gorm:"column:defense;type:jsonb;not null"`
}

func (TeamSeasonPredictedPointsAdded) TableName() string {
   return "cfbd.team_season_predicted_points_added"
}

type TeamGamePredictedPointsAdded struct {
   GameID     int            `gorm:"primaryKey;column:game_id"`
   Team       string         `gorm:"primaryKey;column:team"`
   Season     int            `gorm:"column:season;not null"`
   Week       int            `gorm:"column:week;not null"`
   SeasonType string         `gorm:"column:season_type;not null"`
   Conference string         `gorm:"column:conference;not null"`
   Opponent   string         `gorm:"column:opponent;not null"`
   Offense    datatypes.JSON `gorm:"column:offense;type:jsonb;not null"`
   Defense    datatypes.JSON `gorm:"column:defense;type:jsonb;not null"`
}

func (TeamGamePredictedPointsAdded) TableName() string {
   return "cfbd.team_game_predicted_points_added"
}

type PlayerGamePredictedPointsAdded struct {
   Season     int    `gorm:"primaryKey;column:season"`
   Week       int    `gorm:"primaryKey;column:week"`
   SeasonType string `gorm:"primaryKey;column:season_type"`
   ID         string `gorm:"primaryKey;column:id"`
   Team       string `gorm:"primaryKey;column:team"`

   Name       string         `gorm:"column:name;not null"`
   Position   string         `gorm:"column:position;not null"`
   Opponent   string         `gorm:"column:opponent;not null"`
   AveragePPA datatypes.JSON `gorm:"column:average_ppa;type:jsonb;not null"`
}

func (PlayerGamePredictedPointsAdded) TableName() string {
   return "cfbd.player_game_predicted_points_added"
}

type PlayerSeasonPredictedPointsAdded struct {
   Season     int            `gorm:"primaryKey;column:season"`
   ID         string         `gorm:"primaryKey;column:id"`
   Name       string         `gorm:"column:name;not null"`
   Position   string         `gorm:"column:position;not null"`
   Team       string         `gorm:"column:team;not null"`
   Conference string         `gorm:"column:conference;not null"`
   AveragePPA datatypes.JSON `gorm:"column:average_ppa;type:jsonb;not null"`
   TotalPPA   datatypes.JSON `gorm:"column:total_ppa;type:jsonb;not null"`
}

func (PlayerSeasonPredictedPointsAdded) TableName() string {
   return "cfbd.player_season_predicted_points_added"
}

// ===========================
// Win probability
// ===========================

type PlayWinProbability struct {
   GameID             int     `gorm:"primaryKey;column:game_id"`
   PlayID             string  `gorm:"primaryKey;column:play_id"`
   PlayText           string  `gorm:"column:play_text;not null"`
   HomeID             int     `gorm:"column:home_id;not null"`
   Home               string  `gorm:"column:home;not null"`
   AwayID             int     `gorm:"column:away_id;not null"`
   Away               string  `gorm:"column:away;not null"`
   Spread             float64 `gorm:"column:spread;not null"`
   HomeBall           bool    `gorm:"column:home_ball;not null"`
   HomeScore          int     `gorm:"column:home_score;not null"`
   AwayScore          int     `gorm:"column:away_score;not null"`
   YardLine           int     `gorm:"column:yard_line;not null"`
   Down               int     `gorm:"column:down;not null"`
   Distance           int     `gorm:"column:distance;not null"`
   HomeWinProbability float64 `gorm:"column:home_win_probability;not null"`
   PlayNumber         int     `gorm:"column:play_number;not null"`
}

func (PlayWinProbability) TableName() string { return "cfbd.play_win_probability" }

type PregameWinProbability struct {
   GameID             int     `gorm:"primaryKey;column:game_id"`
   Season             int     `gorm:"column:season;not null"`
   SeasonType         string  `gorm:"column:season_type;not null"`
   Week               int     `gorm:"column:week;not null"`
   HomeTeam           string  `gorm:"column:home_team;not null"`
   AwayTeam           string  `gorm:"column:away_team;not null"`
   Spread             float64 `gorm:"column:spread;not null"`
   HomeWinProbability float64 `gorm:"column:home_win_probability;not null"`
}

func (PregameWinProbability) TableName() string { return "cfbd.pregame_win_probability" }

type FieldGoalEP struct {
   YardsToGoal    int     `gorm:"primaryKey;column:yards_to_goal"`
   Distance       int     `gorm:"primaryKey;column:distance"`
   ExpectedPoints float64 `gorm:"column:expected_points;not null"`
}

func (FieldGoalEP) TableName() string { return "cfbd.field_goal_ep" }

// ===========================
// Live game
// ===========================

type LiveGame struct {
   ID          int    `gorm:"primaryKey;column:id"`
   Status      string `gorm:"column:status;not null"`
   Period      *int   `gorm:"column:period"`
   Clock       string `gorm:"column:clock;not null"`
   Possession  string `gorm:"column:possession;not null"`
   Down        *int   `gorm:"column:down"`
   Distance    *int   `gorm:"column:distance"`
   YardsToGoal *int   `gorm:"column:yards_to_goal"`

   Teams  []LiveGameTeam  `gorm:"foreignKey:LiveGameID;references:ID"`
   Drives []LiveGameDrive `gorm:"foreignKey:LiveGameID;references:ID"`
}

func (LiveGame) TableName() string { return "cfbd.live_games" }

type LiveGameTeam struct {
   LiveGameTeamID          int64    `gorm:"primaryKey;column:live_game_team_id"`
   LiveGameID              int      `gorm:"column:live_game_id;not null;index"`
   TeamID                  int      `gorm:"column:team_id;not null"`
   Team                    string   `gorm:"column:team;not null"`
   HomeAway                string   `gorm:"column:home_away;not null"`
   LineScores              []int    `gorm:"column:line_scores;type:integer[];not null"`
   Points                  int      `gorm:"column:points;not null"`
   Drives                  int      `gorm:"column:drives;not null"`
   ScoringOpportunities    int      `gorm:"column:scoring_opportunities;not null"`
   PointsPerOpportunity    float64  `gorm:"column:points_per_opportunity;not null"`
   AverageStartYardLine    *float64 `gorm:"column:average_start_yard_line"`
   Plays                   int      `gorm:"column:plays;not null"`
   LineYards               float64  `gorm:"column:line_yards;not null"`
   LineYardsPerRush        float64  `gorm:"column:line_yards_per_rush;not null"`
   SecondLevelYards        float64  `gorm:"column:second_level_yards;not null"`
   SecondLevelYardsPerRush float64  `gorm:"column:second_level_yards_per_rush;not null"`
   OpenFieldYards          float64  `gorm:"column:open_field_yards;not null"`
   OpenFieldYardsPerRush   float64  `gorm:"column:open_field_yards_per_rush;not null"`
   EpaPerPlay              float64  `gorm:"column:epa_per_play;not null"`
   TotalEpa                float64  `gorm:"column:total_epa;not null"`
   PassingEpa              float64  `gorm:"column:passing_epa;not null"`
   EpaPerPass              float64  `gorm:"column:epa_per_pass;not null"`
   RushingEpa              float64  `gorm:"column:rushing_epa;not null"`
   EpaPerRush              float64  `gorm:"column:epa_per_rush;not null"`
   SuccessRate             float64  `gorm:"column:success_rate;not null"`
   StandardDownSuccessRate float64  `gorm:"column:standard_down_success_rate;not null"`
   PassingDownSuccessRate  float64  `gorm:"column:passing_down_success_rate;not null"`
   Explosiveness           float64  `gorm:"column:explosiveness;not null"`
   DeserveToWin            *float64 `gorm:"column:deserve_to_win"`

   LiveGame LiveGame `gorm:"foreignKey:LiveGameID;references:ID;constraint:OnDelete:CASCADE"`
}

func (LiveGameTeam) TableName() string { return "cfbd.live_game_teams" }

type LiveGameDrive struct {
   ID                 string  `gorm:"primaryKey;column:id"`
   LiveGameID         int     `gorm:"column:live_game_id;not null;index"`
   OffenseID          int     `gorm:"column:offense_id;not null"`
   Offense            string  `gorm:"column:offense;not null"`
   DefenseID          int     `gorm:"column:defense_id;not null"`
   Defense            string  `gorm:"column:defense;not null"`
   PlayCount          int     `gorm:"column:play_count;not null"`
   Yards              int     `gorm:"column:yards;not null"`
   StartPeriod        int     `gorm:"column:start_period;not null"`
   StartClock         *string `gorm:"column:start_clock"`
   StartYardsToGoal   int     `gorm:"column:start_yards_to_goal;not null"`
   EndPeriod          *int    `gorm:"column:end_period"`
   EndClock           *string `gorm:"column:end_clock"`
   EndYardsToGoal     *int    `gorm:"column:end_yards_to_goal"`
   Duration           *string `gorm:"column:duration"`
   ScoringOpportunity bool    `gorm:"column:scoring_opportunity;not null"`
   Result             string  `gorm:"column:result;not null"`
   PointsGained       int     `gorm:"column:points_gained;not null"`

   Plays []LiveGamePlay `gorm:"foreignKey:DriveID;references:ID"`
}

func (LiveGameDrive) TableName() string { return "cfbd.live_game_drives" }

type LiveGamePlay struct {
   ID          string    `gorm:"primaryKey;column:id"`
   DriveID     string    `gorm:"column:drive_id;not null;index"`
   HomeScore   int       `gorm:"column:home_score;not null"`
   AwayScore   int       `gorm:"column:away_score;not null"`
   Period      int       `gorm:"column:period;not null"`
   Clock       string    `gorm:"column:clock;not null"`
   WallClock   time.Time `gorm:"column:wall_clock;type:timestamptz;not null"`
   TeamID      int       `gorm:"column:team_id;not null"`
   Team        string    `gorm:"column:team;not null"`
   Down        int       `gorm:"column:down;not null"`
   Distance    int       `gorm:"column:distance;not null"`
   YardsToGoal int       `gorm:"column:yards_to_goal;not null"`
   YardsGained int       `gorm:"column:yards_gained;not null"`
   PlayTypeID  int       `gorm:"column:play_type_id;not null"`
   PlayType    string    `gorm:"column:play_type;not null"`
   EPA         *float64  `gorm:"column:epa"`
   GarbageTime bool      `gorm:"column:garbage_time;not null"`
   Success     bool      `gorm:"column:success;not null"`
   RushPass    string    `gorm:"column:rush_pass;not null"`
   DownType    string    `gorm:"column:down_type;not null"`
   PlayText    string    `gorm:"column:play_text;not null"`
}

func (LiveGamePlay) TableName() string { return "cfbd.live_game_plays" }

// ===========================
// Betting / lines / games
// ===========================

type BettingGame struct {
   ID                 int       `gorm:"primaryKey;column:id"`
   Season             int       `gorm:"column:season;not null"`
   SeasonType         string    `gorm:"column:season_type;not null"`
   Week               int       `gorm:"column:week;not null"`
   StartDate          time.Time `gorm:"column:start_date;type:timestamptz;not null"`
   HomeTeamID         int       `gorm:"column:home_team_id;not null"`
   HomeTeam           string    `gorm:"column:home_team;not null"`
   HomeConference     *string   `gorm:"column:home_conference"`
   HomeClassification *string   `gorm:"column:home_classification"`
   HomeScore          *int      `gorm:"column:home_score"`
   AwayTeamID         int       `gorm:"column:away_team_id;not null"`
   AwayTeam           string    `gorm:"column:away_team;not null"`
   AwayConference     *string   `gorm:"column:away_conference"`
   AwayClassification *string   `gorm:"column:away_classification"`
   AwayScore          *int      `gorm:"column:away_score"`

   Lines []GameLine `gorm:"foreignKey:BettingGameID;references:ID"`
}

func (BettingGame) TableName() string { return "cfbd.betting_games" }

type GameLine struct {
   GameLineID      int64    `gorm:"primaryKey;column:game_line_id"`
   BettingGameID   int      `gorm:"column:betting_game_id;not null;index"`
   Provider        string   `gorm:"column:provider;not null"`
   Spread          *float64 `gorm:"column:spread"`
   FormattedSpread *string  `gorm:"column:formatted_spread"`
   SpreadOpen      *float64 `gorm:"column:spread_open"`
   OverUnder       *float64 `gorm:"column:over_under"`
   OverUnderOpen   *float64 `gorm:"column:over_under_open"`
   HomeMoneyline   *float64 `gorm:"column:home_moneyline"`
   AwayMoneyline   *float64 `gorm:"column:away_moneyline"`

   Game BettingGame `gorm:"foreignKey:BettingGameID;references:ID;constraint:OnDelete:CASCADE"`
}

func (GameLine) TableName() string { return "cfbd.game_lines" }

type UserInfo struct {
   PatronLevel    float64 `gorm:"column:patron_level;not null"`
   RemainingCalls float64 `gorm:"column:remaining_calls;not null"`
}

func (UserInfo) TableName() string { return "cfbd.user_info" }

// Games table (main)
type Game struct {
   ID             int       `gorm:"primaryKey;column:id"`
   Season         int       `gorm:"column:season;not null"`
   Week           int       `gorm:"column:week;not null"`
   SeasonType     string    `gorm:"column:season_type;not null"`
   StartDate      time.Time `gorm:"column:start_date;type:timestamptz;not null"`
   StartTimeTBD   bool      `gorm:"column:start_time_tbd;not null"`
   Completed      bool      `gorm:"column:completed;not null"`
   NeutralSite    bool      `gorm:"column:neutral_site;not null"`
   ConferenceGame bool      `gorm:"column:conference_game;not null"`
   Attendance     *int      `gorm:"column:attendance"`
   VenueID        *int      `gorm:"column:venue_id"`
   Venue          *string   `gorm:"column:venue"`

   HomeID                     *int           `gorm:"column:home_id"`
   HomeTeam                   string         `gorm:"column:home_team;not null"`
   HomeConference             *string        `gorm:"column:home_conference"`
   HomeClassification         *string        `gorm:"column:home_classification"`
   HomePoints                 *int           `gorm:"column:home_points"`
   HomeLineScores             datatypes.JSON `gorm:"column:home_line_scores;type:jsonb"`
   HomePostgameWinProbability *float64       `gorm:"column:home_postgame_win_probability"`
   HomePregameElo             *int           `gorm:"column:home_pregame_elo"`
   HomePostgameElo            *int           `gorm:"column:home_postgame_elo"`

   AwayID                     *int           `gorm:"column:away_id"`
   AwayTeam                   string         `gorm:"column:away_team;not null"`
   AwayConference             *string        `gorm:"column:away_conference"`
   AwayClassification         *string        `gorm:"column:away_classification"`
   AwayPoints                 *int           `gorm:"column:away_points"`
   AwayLineScores             datatypes.JSON `gorm:"column:away_line_scores;type:jsonb"`
   AwayPostgameWinProbability *float64       `gorm:"column:away_postgame_win_probability"`
   AwayPregameElo             *int           `gorm:"column:away_pregame_elo"`
   AwayPostgameElo            *int           `gorm:"column:away_postgame_elo"`

   ExcitementIndex *float64 `gorm:"column:excitement_index"`
   Highlights      *string  `gorm:"column:highlights"`
   Notes           *string  `gorm:"column:notes"`
}

func (Game) TableName() string { return "cfbd.games" }

// ===========================
// Game team stats
// ===========================

type GameTeamStats struct {
   ID    int                 `gorm:"primaryKey;column:id"`
   Teams []GameTeamStatsTeam `gorm:"foreignKey:GameID;references:ID"`
}

func (GameTeamStats) TableName() string { return "cfbd.game_team_stats" }

type GameTeamStatsTeam struct {
   GameTeamStatsTeamID int64   `gorm:"primaryKey;column:game_team_stats_team_id"`
   GameID              int     `gorm:"column:game_id;not null;index"`
   TeamID              int     `gorm:"column:team_id;not null"`
   Team                string  `gorm:"column:team;not null"`
   Conference          *string `gorm:"column:conference"`
   HomeAway            string  `gorm:"column:home_away;not null"`
   Points              *int    `gorm:"column:points"`

   Stats []GameTeamStatsTeamStat `gorm:"foreignKey:GameTeamStatsTeamID;references:GameTeamStatsTeamID"`
}

func (GameTeamStatsTeam) TableName() string { return "cfbd.game_team_stats_teams" }

type GameTeamStatsTeamStat struct {
   GameTeamStatsTeamStatID int64  `gorm:"primaryKey;column:game_team_stats_team_stat_id"`
   GameTeamStatsTeamID     int64  `gorm:"column:game_team_stats_team_id;not null;index"`
   Category                string `gorm:"column:category;not null"`
   Stat                    string `gorm:"column:stat;not null"`
}

func (GameTeamStatsTeamStat) TableName() string { return "cfbd.game_team_stats_team_stats" }

// ===========================
// Game player stats (JSONB)
// ===========================

type GamePlayerStats struct {
   ID    int            `gorm:"primaryKey;column:id"`
   Teams datatypes.JSON `gorm:"column:teams;type:jsonb;not null"`
}

func (GamePlayerStats) TableName() string { return "cfbd.game_player_stats" }

// ===========================
// Media & weather
// ===========================

type GameMedia struct {
   ID             int       `gorm:"primaryKey;column:id"`
   Season         int       `gorm:"column:season;not null"`
   Week           int       `gorm:"column:week;not null"`
   SeasonType     string    `gorm:"column:season_type;not null"`
   StartTime      time.Time `gorm:"column:start_time;type:timestamptz;not null"`
   IsStartTimeTBD bool      `gorm:"column:is_start_time_tbd;not null"`
   HomeTeam       string    `gorm:"column:home_team;not null"`
   HomeConference *string   `gorm:"column:home_conference"`
   AwayTeam       string    `gorm:"column:away_team;not null"`
   AwayConference *string   `gorm:"column:away_conference"`
   MediaType      string    `gorm:"column:media_type;not null"`
   Outlet         string    `gorm:"column:outlet;not null"`
}

func (GameMedia) TableName() string { return "cfbd.game_media" }

type GameWeather struct {
   ID             int       `gorm:"primaryKey;column:id"`
   Season         int       `gorm:"column:season;not null"`
   Week           int       `gorm:"column:week;not null"`
   SeasonType     string    `gorm:"column:season_type;not null"`
   StartTime      time.Time `gorm:"column:start_time;type:timestamptz;not null"`
   GameIndoors    bool      `gorm:"column:game_indoors;not null"`
   HomeTeam       string    `gorm:"column:home_team;not null"`
   HomeConference *string   `gorm:"column:home_conference"`
   AwayTeam       string    `gorm:"column:away_team;not null"`
   AwayConference *string   `gorm:"column:away_conference"`
   VenueID        *int      `gorm:"column:venue_id"`
   Venue          *string   `gorm:"column:venue"`

   Temperature          *float64 `gorm:"column:temperature"`
   DewPoint             *float64 `gorm:"column:dew_point"`
   Humidity             *float64 `gorm:"column:humidity"`
   Precipitation        *float64 `gorm:"column:precipitation"`
   Snowfall             *float64 `gorm:"column:snowfall"`
   WindDirection        *float64 `gorm:"column:wind_direction"`
   WindSpeed            *float64 `gorm:"column:wind_speed"`
   Pressure             *float64 `gorm:"column:pressure"`
   WeatherConditionCode *float64 `gorm:"column:weather_condition_code"`
   WeatherCondition     *string  `gorm:"column:weather_condition"`
}

func (GameWeather) TableName() string { return "cfbd.game_weather" }

// ===========================
// Records / calendar / scoreboard
// ===========================

type TeamRecords struct {
   Year           int      `gorm:"primaryKey;column:year"`
   Team           string   `gorm:"primaryKey;column:team"`
   TeamID         *int     `gorm:"column:team_id"`
   Classification *string  `gorm:"column:classification"`
   Conference     string   `gorm:"column:conference;not null"`
   Division       string   `gorm:"column:division;not null"`
   ExpectedWins   *float64 `gorm:"column:expected_wins"`

   Total            datatypes.JSON `gorm:"column:total;type:jsonb;not null"`
   ConferenceGames  datatypes.JSON `gorm:"column:conference_games;type:jsonb;not null"`
   HomeGames        datatypes.JSON `gorm:"column:home_games;type:jsonb;not null"`
   AwayGames        datatypes.JSON `gorm:"column:away_games;type:jsonb;not null"`
   NeutralSiteGames datatypes.JSON `gorm:"column:neutral_site_games;type:jsonb;not null"`
   RegularSeason    datatypes.JSON `gorm:"column:regular_season;type:jsonb;not null"`
   Postseason       datatypes.JSON `gorm:"column:postseason;type:jsonb;not null"`
}

func (TeamRecords) TableName() string { return "cfbd.team_records" }

type CalendarWeek struct {
   Season         int        `gorm:"primaryKey;column:season"`
   SeasonType     string     `gorm:"primaryKey;column:season_type"`
   Week           int        `gorm:"primaryKey;column:week"`
   StartDate      time.Time  `gorm:"column:start_date;type:timestamptz;not null"`
   EndDate        time.Time  `gorm:"column:end_date;type:timestamptz;not null"`
   FirstGameStart *time.Time `gorm:"column:first_game_start;type:timestamptz"`
   LastGameStart  *time.Time `gorm:"column:last_game_start;type:timestamptz"`
}

func (CalendarWeek) TableName() string { return "cfbd.calendar_weeks" }

type Scoreboard struct {
   ID             int       `gorm:"primaryKey;column:id"`
   StartDate      time.Time `gorm:"column:start_date;type:timestamptz;not null"`
   StartTimeTBD   bool      `gorm:"column:start_time_tbd;not null"`
   TV             *string   `gorm:"column:tv"`
   NeutralSite    bool      `gorm:"column:neutral_site;not null"`
   ConferenceGame bool      `gorm:"column:conference_game;not null"`
   Status         string    `gorm:"column:status;not null"`
   Period         *int      `gorm:"column:period"`
   Clock          *string   `gorm:"column:clock"`
   Situation      *string   `gorm:"column:situation"`
   Possession     *string   `gorm:"column:possession"`
   LastPlay       *string   `gorm:"column:last_play"`

   Venue    datatypes.JSON `gorm:"column:venue;type:jsonb;not null"`
   HomeTeam datatypes.JSON `gorm:"column:home_team;type:jsonb;not null"`
   AwayTeam datatypes.JSON `gorm:"column:away_team;type:jsonb;not null"`
   Weather  datatypes.JSON `gorm:"column:weather;type:jsonb;not null"`
   Betting  datatypes.JSON `gorm:"column:betting;type:jsonb;not null"`
}

func (Scoreboard) TableName() string { return "cfbd.scoreboards" }

// ===========================
// Drives
// ===========================

type Drive struct {
   ID                string  `gorm:"primaryKey;column:id"`
   GameID            int     `gorm:"column:game_id;not null;index"`
   Offense           string  `gorm:"column:offense;not null"`
   OffenseConference *string `gorm:"column:offense_conference"`
   Defense           string  `gorm:"column:defense;not null"`
   DefenseConference *string `gorm:"column:defense_conference"`
   DriveNumber       *int    `gorm:"column:drive_number"`
   Scoring           bool    `gorm:"column:scoring;not null"`
   StartPeriod       int     `gorm:"column:start_period;not null"`
   StartYardline     int     `gorm:"column:start_yardline;not null"`
   StartYardsToGoal  int     `gorm:"column:start_yards_to_goal;not null"`
   StartTimeSeconds  *int    `gorm:"column:start_time_seconds"`
   StartTimeMinutes  *int    `gorm:"column:start_time_minutes"`
   EndPeriod         int     `gorm:"column:end_period;not null"`
   EndYardline       int     `gorm:"column:end_yardline;not null"`
   EndYardsToGoal    int     `gorm:"column:end_yards_to_goal;not null"`
   EndTimeSeconds    *int    `gorm:"column:end_time_seconds"`
   EndTimeMinutes    *int    `gorm:"column:end_time_minutes"`
   ElapsedSeconds    *int    `gorm:"column:elapsed_seconds"`
   ElapsedMinutes    *int    `gorm:"column:elapsed_minutes"`
   Plays             int     `gorm:"column:plays;not null"`
   Yards             int     `gorm:"column:yards;not null"`
   DriveResult       string  `gorm:"column:drive_result;not null"`
   IsHomeOffense     bool    `gorm:"column:is_home_offense;not null"`
   StartOffenseScore int     `gorm:"column:start_offense_score;not null"`
   StartDefenseScore int     `gorm:"column:start_defense_score;not null"`
   EndOffenseScore   int     `gorm:"column:end_offense_score;not null"`
   EndDefenseScore   int     `gorm:"column:end_defense_score;not null"`
}

func (Drive) TableName() string { return "cfbd.drives" }

// ===========================
// Draft
// ===========================

type DraftPick struct {
   DraftPickID int64 `gorm:"primaryKey;column:draft_pick_id"`

   CollegeAthleteID *int `gorm:"column:college_athlete_id"`
   NFIAthleteID     *int `gorm:"column:nfl_athlete_id"`

   CollegeID         int     `gorm:"column:college_id;not null"`
   CollegeTeam       string  `gorm:"column:college_team;not null"`
   CollegeConference *string `gorm:"column:college_conference"`

   NFLTeamID int    `gorm:"column:nfl_team_id;not null"`
   NFLTeam   string `gorm:"column:nfl_team;not null"`

   Year    int `gorm:"column:year;not null;index"`
   Overall int `gorm:"column:overall;not null"`
   Round   int `gorm:"column:round;not null"`
   Pick    int `gorm:"column:pick;not null"`

   Name     string `gorm:"column:name;not null"`
   Position string `gorm:"column:position;not null"`

   Height *float64 `gorm:"column:height"`
   Weight *int     `gorm:"column:weight"`

   PreDraftRanking         *int `gorm:"column:pre_draft_ranking"`
   PreDraftPositionRanking *int `gorm:"column:pre_draft_position_ranking"`
   PreDraftGrade           *int `gorm:"column:pre_draft_grade"`

   HometownCountyFips *string `gorm:"column:hometown_county_fips"`
   HometownLongitude  *string `gorm:"column:hometown_longitude"`
   HometownLatitude   *string `gorm:"column:hometown_latitude"`
   HometownCountry    *string `gorm:"column:hometown_country"`
   HometownState      *string `gorm:"column:hometown_state"`
   HometownCity       *string `gorm:"column:hometown_city"`
}

func (DraftPick) TableName() string { return "cfbd.draft_picks" }

// ===========================
// Coaches
// ===========================

type Coach struct {
   CoachID   int64     `gorm:"primaryKey;column:coach_id"`
   FirstName string    `gorm:"column:first_name;not null"`
   LastName  string    `gorm:"column:last_name;not null"`
   HireDate  time.Time `gorm:"column:hire_date;type:timestamptz;not null"`

   Seasons []CoachSeason `gorm:"foreignKey:CoachID;references:CoachID"`
}

func (Coach) TableName() string { return "cfbd.coaches" }

type CoachSeason struct {
   CoachSeasonID int64 `gorm:"primaryKey;column:coach_season_id"`
   CoachID       int64 `gorm:"column:coach_id;not null;index"`

   School string `gorm:"column:school;not null"`
   Year   int    `gorm:"column:year;not null"`

   Games  int `gorm:"column:games;not null"`
   Wins   int `gorm:"column:wins;not null"`
   Losses int `gorm:"column:losses;not null"`
   Ties   int `gorm:"column:ties;not null"`

   PreseasonRank  *int     `gorm:"column:preseason_rank"`
   PostseasonRank *int     `gorm:"column:postseason_rank"`
   SRS            *float64 `gorm:"column:srs"`
   SPOverall      *float64 `gorm:"column:sp_overall"`
   SPOffense      *float64 `gorm:"column:sp_offense"`
   SPDefense      *float64 `gorm:"column:sp_defense"`

   Coach Coach `gorm:"foreignKey:CoachID;references:CoachID;constraint:OnDelete:CASCADE"`
}

func (CoachSeason) TableName() string { return "cfbd.coach_seasons" }

// ===========================
// Recruiting
// ===========================

type Recruit struct {
   ID            string   `gorm:"primaryKey;column:id"`
   AthleteID     *string  `gorm:"column:athlete_id"`
   RecruitType   string   `gorm:"column:recruit_type;not null"`
   Year          int      `gorm:"column:year;not null"`
   Ranking       *int     `gorm:"column:ranking"`
   Name          string   `gorm:"column:name;not null"`
   School        *string  `gorm:"column:school"`
   CommittedTo   *string  `gorm:"column:committed_to"`
   Position      *string  `gorm:"column:position"`
   Height        *float64 `gorm:"column:height"`
   Weight        *int     `gorm:"column:weight"`
   Stars         int      `gorm:"column:stars;not null"`
   Rating        float64  `gorm:"column:rating;not null"`
   City          *string  `gorm:"column:city"`
   StateProvince *string  `gorm:"column:state_province"`
   Country       *string  `gorm:"column:country"`

   HometownFipsCode  *string  `gorm:"column:hometown_fips_code"`
   HometownLongitude *float64 `gorm:"column:hometown_longitude"`
   HometownLatitude  *float64 `gorm:"column:hometown_latitude"`
}

func (Recruit) TableName() string { return "cfbd.recruits" }

type TeamRecruitingRanking struct {
   Year   int     `gorm:"primaryKey;column:year"`
   Team   string  `gorm:"primaryKey;column:team"`
   Rank   int     `gorm:"column:rank;not null"`
   Points float64 `gorm:"column:points;not null"`
}

func (TeamRecruitingRanking) TableName() string { return "cfbd.team_recruiting_rankings" }

type AggregatedTeamRecruiting struct {
   Team          string  `gorm:"primaryKey;column:team"`
   Conference    string  `gorm:"primaryKey;column:conference"`
   PositionGroup *string `gorm:"primaryKey;column:position_group"`
   AverageRating float64 `gorm:"column:average_rating;not null"`
   TotalRating   float64 `gorm:"column:total_rating;not null"`
   Commits       int     `gorm:"column:commits;not null"`
   AverageStars  float64 `gorm:"column:average_stars;not null"`
}

func (AggregatedTeamRecruiting) TableName() string { return "cfbd.aggregated_team_recruiting" }

// ===========================
// Game havoc stats (jsonb sides)
// ===========================

type GameHavocStats struct {
   GameID             int            `gorm:"primaryKey;column:game_id"`
   Team               string         `gorm:"primaryKey;column:team"`
   Season             int            `gorm:"column:season;not null"`
   SeasonType         string         `gorm:"column:season_type;not null"`
   Week               int            `gorm:"column:week;not null"`
   Conference         *string        `gorm:"column:conference"`
   Opponent           string         `gorm:"column:opponent;not null"`
   OpponentConference *string        `gorm:"column:opponent_conference"`
   Offense            datatypes.JSON `gorm:"column:offense;type:jsonb;not null"`
   Defense            datatypes.JSON `gorm:"column:defense;type:jsonb;not null"`
}

func (GameHavocStats) TableName() string { return "cfbd.game_havoc_stats" }

// ===========================
// Advanced Season Stats (normalized)
// ===========================

type AdvRateMetrics struct {
   ID            int64    `gorm:"primaryKey;column:adv_rate_metrics_id"`
   Explosiveness *float64 `gorm:"column:explosiveness"`
   SuccessRate   *float64 `gorm:"column:success_rate"`
   TotalPPA      *float64 `gorm:"column:total_ppa"`
   PPA           *float64 `gorm:"column:ppa"`
   Rate          *float64 `gorm:"column:rate"`
}

func (AdvRateMetrics) TableName() string { return "cfbd.adv_rate_metrics" }

type AdvHavoc struct {
   ID         int64    `gorm:"primaryKey;column:adv_havoc_id"`
   DB         *float64 `gorm:"column:db"`
   FrontSeven *float64 `gorm:"column:front_seven"`
   Total      *float64 `gorm:"column:total"`
}

func (AdvHavoc) TableName() string { return "cfbd.adv_havoc" }

type AdvFieldPosition struct {
   ID                     int64    `gorm:"primaryKey;column:adv_field_position_id"`
   AveragePredictedPoints *float64 `gorm:"column:average_predicted_points"`
   AverageStart           *float64 `gorm:"column:average_start"`
}

func (AdvFieldPosition) TableName() string { return "cfbd.adv_field_position" }

type AdvSeasonStatSide struct {
   ID int64 `gorm:"primaryKey;column:adv_season_stat_side_id"`

   PassingPlaysID  *int64 `gorm:"column:passing_plays_id"`
   RushingPlaysID  *int64 `gorm:"column:rushing_plays_id"`
   PassingDownsID  *int64 `gorm:"column:passing_downs_id"`
   StandardDownsID *int64 `gorm:"column:standard_downs_id"`
   HavocID         *int64 `gorm:"column:havoc_id"`
   FieldPositionID *int64 `gorm:"column:field_position_id"`

   PointsPerOpportunity  *float64 `gorm:"column:points_per_opportunity"`
   TotalOpportunies      *int     `gorm:"column:total_opportunies"`
   OpenFieldYardsTotal   *int     `gorm:"column:open_field_yards_total"`
   OpenFieldYards        *float64 `gorm:"column:open_field_yards"`
   SecondLevelYardsTotal *int     `gorm:"column:second_level_yards_total"`
   SecondLevelYards      *float64 `gorm:"column:second_level_yards"`
   LineYardsTotal        *int     `gorm:"column:line_yards_total"`
   LineYards             *float64 `gorm:"column:line_yards"`
   StuffRate             *float64 `gorm:"column:stuff_rate"`
   PowerSuccess          *float64 `gorm:"column:power_success"`

   Explosiveness *float64 `gorm:"column:explosiveness"`
   SuccessRate   *float64 `gorm:"column:success_rate"`
   TotalPPA      *float64 `gorm:"column:total_ppa"`
   PPA           *float64 `gorm:"column:ppa"`

   Drives *int `gorm:"column:drives"`
   Plays  *int `gorm:"column:plays"`

   PassingPlays  *AdvRateMetrics   `gorm:"foreignKey:PassingPlaysID;references:ID;constraint:OnDelete:SET NULL"`
   RushingPlays  *AdvRateMetrics   `gorm:"foreignKey:RushingPlaysID;references:ID;constraint:OnDelete:SET NULL"`
   PassingDowns  *AdvRateMetrics   `gorm:"foreignKey:PassingDownsID;references:ID;constraint:OnDelete:SET NULL"`
   StandardDowns *AdvRateMetrics   `gorm:"foreignKey:StandardDownsID;references:ID;constraint:OnDelete:SET NULL"`
   Havoc         *AdvHavoc         `gorm:"foreignKey:HavocID;references:ID;constraint:OnDelete:SET NULL"`
   FieldPosition *AdvFieldPosition `gorm:"foreignKey:FieldPositionID;references:ID;constraint:OnDelete:SET NULL"`
}

func (AdvSeasonStatSide) TableName() string { return "cfbd.adv_season_stat_side" }

type AdvancedSeasonStatsNormalized struct {
   Season     int    `gorm:"primaryKey;column:season"`
   Team       string `gorm:"primaryKey;column:team"`
   Conference string `gorm:"column:conference;not null"`

   OffenseSideID int64 `gorm:"column:offense_side_id;not null;index"`
   DefenseSideID int64 `gorm:"column:defense_side_id;not null;index"`

   Offense *AdvSeasonStatSide `gorm:"foreignKey:OffenseSideID;references:ID;constraint:OnDelete:RESTRICT"`
   Defense *AdvSeasonStatSide `gorm:"foreignKey:DefenseSideID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (AdvancedSeasonStatsNormalized) TableName() string {
   return "cfbd.advanced_season_stats_normalized"
}

// ===========================
// Advanced Game Stats (normalized)
// ===========================

type AdvGamePlayMetrics struct {
   ID            int64    `gorm:"primaryKey;column:adv_game_play_metrics_id"`
   Explosiveness *float64 `gorm:"column:explosiveness"`
   SuccessRate   *float64 `gorm:"column:success_rate"`
   TotalPPA      *float64 `gorm:"column:total_ppa"`
   PPA           *float64 `gorm:"column:ppa"`
}

func (AdvGamePlayMetrics) TableName() string { return "cfbd.adv_game_play_metrics" }

type AdvGameDownMetrics struct {
   ID            int64    `gorm:"primaryKey;column:adv_game_down_metrics_id"`
   Explosiveness *float64 `gorm:"column:explosiveness"`
   SuccessRate   *float64 `gorm:"column:success_rate"`
   PPA           *float64 `gorm:"column:ppa"`
}

func (AdvGameDownMetrics) TableName() string { return "cfbd.adv_game_down_metrics" }

type AdvGameStatSide struct {
   ID int64 `gorm:"primaryKey;column:adv_game_stat_side_id"`

   PassingPlaysID  *int64 `gorm:"column:passing_plays_id"`
   RushingPlaysID  *int64 `gorm:"column:rushing_plays_id"`
   PassingDownsID  *int64 `gorm:"column:passing_downs_id"`
   StandardDownsID *int64 `gorm:"column:standard_downs_id"`

   OpenFieldYardsTotal   *int     `gorm:"column:open_field_yards_total"`
   OpenFieldYards        *float64 `gorm:"column:open_field_yards"`
   SecondLevelYardsTotal *int     `gorm:"column:second_level_yards_total"`
   SecondLevelYards      *float64 `gorm:"column:second_level_yards"`
   LineYardsTotal        *int     `gorm:"column:line_yards_total"`
   LineYards             *float64 `gorm:"column:line_yards"`

   StuffRate    *float64 `gorm:"column:stuff_rate"`
   PowerSuccess *float64 `gorm:"column:power_success"`

   Explosiveness *float64 `gorm:"column:explosiveness"`
   SuccessRate   *float64 `gorm:"column:success_rate"`
   TotalPPA      *float64 `gorm:"column:total_ppa"`
   PPA           *float64 `gorm:"column:ppa"`

   Drives *int `gorm:"column:drives"`
   Plays  *int `gorm:"column:plays"`

   PassingPlays  *AdvGamePlayMetrics `gorm:"foreignKey:PassingPlaysID;references:ID;constraint:OnDelete:SET NULL"`
   RushingPlays  *AdvGamePlayMetrics `gorm:"foreignKey:RushingPlaysID;references:ID;constraint:OnDelete:SET NULL"`
   PassingDowns  *AdvGameDownMetrics `gorm:"foreignKey:PassingDownsID;references:ID;constraint:OnDelete:SET NULL"`
   StandardDowns *AdvGameDownMetrics `gorm:"foreignKey:StandardDownsID;references:ID;constraint:OnDelete:SET NULL"`
}

func (AdvGameStatSide) TableName() string { return "cfbd.adv_game_stat_side" }

type AdvancedGameStatsNormalized struct {
   GameID int    `gorm:"primaryKey;column:game_id"`
   Team   string `gorm:"primaryKey;column:team"`

   Season     int    `gorm:"column:season;not null"`
   SeasonType string `gorm:"column:season_type;not null"`
   Week       int    `gorm:"column:week;not null"`
   Opponent   string `gorm:"column:opponent;not null"`

   OffenseSideID int64 `gorm:"column:offense_side_id;not null;index"`
   DefenseSideID int64 `gorm:"column:defense_side_id;not null;index"`

   Offense *AdvGameStatSide `gorm:"foreignKey:OffenseSideID;references:ID;constraint:OnDelete:RESTRICT"`
   Defense *AdvGameStatSide `gorm:"foreignKey:DefenseSideID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (AdvancedGameStatsNormalized) TableName() string {
   return "cfbd.advanced_game_stats_normalized"
}

// ===========================
// Advanced Box Score (normalized)
// ===========================

type AdvancedBoxScoreGameInfo struct {
   ID          int64   `gorm:"primaryKey;column:abs_game_info_id"`
   Excitement  float64 `gorm:"column:excitement;not null"`
   HomeWinner  bool    `gorm:"column:home_winner;not null"`
   AwayWinProb float64 `gorm:"column:away_win_prob;not null"`
   AwayPoints  int     `gorm:"column:away_points;not null"`
   AwayTeam    string  `gorm:"column:away_team;not null"`
   HomeWinProb float64 `gorm:"column:home_win_prob;not null"`
   HomePoints  int     `gorm:"column:home_points;not null"`
   HomeTeam    string  `gorm:"column:home_team;not null"`
}

func (AdvancedBoxScoreGameInfo) TableName() string { return "cfbd.advanced_box_score_game_info" }

type AdvancedBoxScore struct {
   AdvancedBoxScoreID int64 `gorm:"primaryKey;column:advanced_box_score_id"`
   GameInfoID         int64 `gorm:"column:game_info_id;not null"`

   GameInfo AdvancedBoxScoreGameInfo `gorm:"foreignKey:GameInfoID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (AdvancedBoxScore) TableName() string { return "cfbd.advanced_box_scores" }

type StatsByQuarter struct {
   ID       int64    `gorm:"primaryKey;column:stats_by_quarter_id"`
   Total    float64  `gorm:"column:total;not null"`
   Quarter1 *float64 `gorm:"column:quarter1"`
   Quarter2 *float64 `gorm:"column:quarter2"`
   Quarter3 *float64 `gorm:"column:quarter3"`
   Quarter4 *float64 `gorm:"column:quarter4"`
}

func (StatsByQuarter) TableName() string { return "cfbd.stats_by_quarter" }

type AbsTeamFieldPosition struct {
   ID                             int64   `gorm:"primaryKey;column:abs_team_field_position_id"`
   AdvancedBoxScoreID             int64   `gorm:"column:advanced_box_score_id;not null;index"`
   Team                           string  `gorm:"column:team;not null"`
   AverageStart                   float64 `gorm:"column:average_start;not null"`
   AverageStartingPredictedPoints float64 `gorm:"column:average_starting_predicted_points;not null"`
}

func (AbsTeamFieldPosition) TableName() string { return "cfbd.abs_team_field_position" }

type AbsTeamScoringOpportunities struct {
   ID                   int64   `gorm:"primaryKey;column:abs_team_scoring_opportunities_id"`
   AdvancedBoxScoreID   int64   `gorm:"column:advanced_box_score_id;not null;index"`
   Team                 string  `gorm:"column:team;not null"`
   Opportunities        int     `gorm:"column:opportunities;not null"`
   Points               int     `gorm:"column:points;not null"`
   PointsPerOpportunity float64 `gorm:"column:points_per_opportunity;not null"`
}

func (AbsTeamScoringOpportunities) TableName() string { return "cfbd.abs_team_scoring_opportunities" }

type AbsTeamHavoc struct {
   ID                 int64   `gorm:"primaryKey;column:abs_team_havoc_id"`
   AdvancedBoxScoreID int64   `gorm:"column:advanced_box_score_id;not null;index"`
   Team               string  `gorm:"column:team;not null"`
   Total              float64 `gorm:"column:total;not null"`
   FrontSeven         float64 `gorm:"column:front_seven;not null"`
   DB                 float64 `gorm:"column:db;not null"`
}

func (AbsTeamHavoc) TableName() string { return "cfbd.abs_team_havoc" }

type AbsTeamRushingStats struct {
   ID                      int64   `gorm:"primaryKey;column:abs_team_rushing_stats_id"`
   AdvancedBoxScoreID      int64   `gorm:"column:advanced_box_score_id;not null;index"`
   Team                    string  `gorm:"column:team;not null"`
   PowerSuccess            float64 `gorm:"column:power_success;not null"`
   StuffRate               float64 `gorm:"column:stuff_rate;not null"`
   LineYards               float64 `gorm:"column:line_yards;not null"`
   LineYardsAverage        float64 `gorm:"column:line_yards_average;not null"`
   SecondLevelYards        float64 `gorm:"column:second_level_yards;not null"`
   SecondLevelYardsAverage float64 `gorm:"column:second_level_yards_average;not null"`
   OpenFieldYards          float64 `gorm:"column:open_field_yards;not null"`
   OpenFieldYardsAverage   float64 `gorm:"column:open_field_yards_average;not null"`
}

func (AbsTeamRushingStats) TableName() string { return "cfbd.abs_team_rushing_stats" }

type AbsTeamExplosiveness struct {
   ID                 int64  `gorm:"primaryKey;column:abs_team_explosiveness_id"`
   AdvancedBoxScoreID int64  `gorm:"column:advanced_box_score_id;not null;index"`
   Team               string `gorm:"column:team;not null"`
   OverallID          int64  `gorm:"column:overall_id;not null"`

   Overall StatsByQuarter `gorm:"foreignKey:OverallID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (AbsTeamExplosiveness) TableName() string { return "cfbd.abs_team_explosiveness" }

type AbsTeamSuccessRates struct {
   ID                 int64  `gorm:"primaryKey;column:abs_team_success_rates_id"`
   AdvancedBoxScoreID int64  `gorm:"column:advanced_box_score_id;not null;index"`
   Team               string `gorm:"column:team;not null"`

   OverallID       int64 `gorm:"column:overall_id;not null"`
   StandardDownsID int64 `gorm:"column:standard_downs_id;not null"`
   PassingDownsID  int64 `gorm:"column:passing_downs_id;not null"`

   Overall       StatsByQuarter `gorm:"foreignKey:OverallID;references:ID;constraint:OnDelete:RESTRICT"`
   StandardDowns StatsByQuarter `gorm:"foreignKey:StandardDownsID;references:ID;constraint:OnDelete:RESTRICT"`
   PassingDowns  StatsByQuarter `gorm:"foreignKey:PassingDownsID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (AbsTeamSuccessRates) TableName() string { return "cfbd.abs_team_success_rates" }

type AbsTeamPPA struct {
   ID                 int64  `gorm:"primaryKey;column:abs_team_ppa_id"`
   AdvancedBoxScoreID int64  `gorm:"column:advanced_box_score_id;not null;index"`
   Team               string `gorm:"column:team;not null"`
   Plays              int    `gorm:"column:plays;not null"`

   OverallID int64 `gorm:"column:overall_id;not null"`
   PassingID int64 `gorm:"column:passing_id;not null"`
   RushingID int64 `gorm:"column:rushing_id;not null"`

   Kind string `gorm:"column:kind;not null;index"` // add CHECK via db.Exec if desired

   Overall StatsByQuarter `gorm:"foreignKey:OverallID;references:ID;constraint:OnDelete:RESTRICT"`
   Passing StatsByQuarter `gorm:"foreignKey:PassingID;references:ID;constraint:OnDelete:RESTRICT"`
   Rushing StatsByQuarter `gorm:"foreignKey:RushingID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (AbsTeamPPA) TableName() string { return "cfbd.abs_team_ppa" }

type PlayerStatsByQuarter struct {
   ID       int64    `gorm:"primaryKey;column:player_stats_by_quarter_id"`
   Total    float64  `gorm:"column:total;not null"`
   Quarter1 *float64 `gorm:"column:quarter1"`
   Quarter2 *float64 `gorm:"column:quarter2"`
   Quarter3 *float64 `gorm:"column:quarter3"`
   Quarter4 *float64 `gorm:"column:quarter4"`
   Rushing  float64  `gorm:"column:rushing;not null"`
   Passing  float64  `gorm:"column:passing;not null"`
}

func (PlayerStatsByQuarter) TableName() string { return "cfbd.player_stats_by_quarter" }

type AbsPlayerPPA struct {
   ID                 int64  `gorm:"primaryKey;column:abs_player_ppa_id"`
   AdvancedBoxScoreID int64  `gorm:"column:advanced_box_score_id;not null;index"`
   Player             string `gorm:"column:player;not null"`
   Team               string `gorm:"column:team;not null"`
   Position           string `gorm:"column:position;not null"`

   AverageID    int64 `gorm:"column:average_id;not null"`
   CumulativeID int64 `gorm:"column:cumulative_id;not null"`

   Average    PlayerStatsByQuarter `gorm:"foreignKey:AverageID;references:ID;constraint:OnDelete:RESTRICT"`
   Cumulative PlayerStatsByQuarter `gorm:"foreignKey:CumulativeID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (AbsPlayerPPA) TableName() string { return "cfbd.abs_player_ppa" }

type PlayerGameUsageQuarters struct {
   ID       int64    `gorm:"primaryKey;column:player_game_usage_quarters_id"`
   Total    float64  `gorm:"column:total;not null"`
   Quarter1 *float64 `gorm:"column:quarter1"`
   Quarter2 *float64 `gorm:"column:quarter2"`
   Quarter3 *float64 `gorm:"column:quarter3"`
   Quarter4 *float64 `gorm:"column:quarter4"`
   Rushing  float64  `gorm:"column:rushing;not null"`
   Passing  float64  `gorm:"column:passing;not null"`
}

func (PlayerGameUsageQuarters) TableName() string { return "cfbd.player_game_usage_quarters" }

type AbsPlayerGameUsage struct {
   ID                 int64  `gorm:"primaryKey;column:abs_player_game_usage_id"`
   AdvancedBoxScoreID int64  `gorm:"column:advanced_box_score_id;not null;index"`
   Player             string `gorm:"column:player;not null"`
   Team               string `gorm:"column:team;not null"`
   Position           string `gorm:"column:position;not null"`

   UsageQuartersID int64                   `gorm:"column:usage_quarters_id;not null"`
   UsageQuarters   PlayerGameUsageQuarters `gorm:"foreignKey:UsageQuartersID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (AbsPlayerGameUsage) TableName() string { return "cfbd.abs_player_game_usage" }

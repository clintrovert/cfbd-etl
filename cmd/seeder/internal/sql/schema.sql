-- =====================================================================
-- Best-effort relational mapping of cfbd.v1 protobuf messages to Postgres
-- Includes:
--   - Core/reference entities (venues, teams, conferences, games, etc.)
--   - Metrics tables
--   - Plays/drives/betting/media/weather/records
--   - Normalized AdvancedSeasonStat / AdvancedGameStat
--   - Fully normalized AdvancedBoxScore
--
-- Conventions:
--   * proto scalar -> SQL scalar
--   * google.protobuf.*Value wrappers -> nullable columns
--   * google.protobuf.Value / Struct / ListValue -> jsonb
--   * repeated -> child tables
--   * some messages are "API response shapes" that are deeply nested;
--     for those we either normalize or (when not requested) store jsonb.
-- =====================================================================

BEGIN;

CREATE SCHEMA IF NOT EXISTS cfbd;

-- Optional: keep everything in this schema by default
-- SET search_path TO cfbd, public;

-- =====================================================================
-- Core reference tables
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.venues (
  id                INTEGER PRIMARY KEY, -- Venue.id (nullable in proto, but PK when present)
  name              TEXT,
  city              TEXT,
  state             TEXT,
  zip               TEXT,
  country_code      TEXT,
  timezone          TEXT,
  latitude          DOUBLE PRECISION,
  longitude         DOUBLE PRECISION,
  elevation         TEXT,
  capacity          INTEGER,
  construction_year INTEGER,
  grass             BOOLEAN,
  dome              BOOLEAN
);

CREATE TABLE IF NOT EXISTS cfbd.teams (
  id                INTEGER PRIMARY KEY, -- Team.id
  school            TEXT NOT NULL,
  mascot            TEXT,
  abbreviation      TEXT,
  alternate_names   JSONB, -- google.protobuf.ListValue
  conference        TEXT,
  division          TEXT,
  classification    TEXT,
  color             TEXT,
  alternate_color   TEXT,
  logos             JSONB,-- google.protobuf.ListValue
  twitter           TEXT NOT NULL,
  venue_id          INTEGER REFERENCES cfbd.venues(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS cfbd.conferences (
  id              INTEGER PRIMARY KEY,           -- Conference.id
  name            TEXT NOT NULL,
  short_name      TEXT,
  abbreviation    TEXT,
  classification  TEXT
);

-- =====================================================================
-- Simple top-level metrics tables
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.adjusted_team_metrics (
  year                      INTEGER NOT NULL,
  team_id                   INTEGER NOT NULL,
  team                      TEXT NOT NULL,
  conference                TEXT NOT NULL,

  -- EpaSplit epa
  epa_rushing               DOUBLE PRECISION NOT NULL,
  epa_passing               DOUBLE PRECISION NOT NULL,
  epa_total                 DOUBLE PRECISION NOT NULL,

  -- EpaSplit epa_allowed
  epa_allowed_rushing       DOUBLE PRECISION NOT NULL,
  epa_allowed_passing       DOUBLE PRECISION NOT NULL,
  epa_allowed_total         DOUBLE PRECISION NOT NULL,

  -- SuccessRateSplit success_rate
  success_rate_passing_downs   DOUBLE PRECISION NOT NULL,
  success_rate_standard_downs  DOUBLE PRECISION NOT NULL,
  success_rate_total           DOUBLE PRECISION NOT NULL,

  -- SuccessRateSplit success_rate_allowed
  success_rate_allowed_passing_downs   DOUBLE PRECISION NOT NULL,
  success_rate_allowed_standard_downs  DOUBLE PRECISION NOT NULL,
  success_rate_allowed_total           DOUBLE PRECISION NOT NULL,

  -- RushingYardsSplit rushing
  rushing_highlight_yards    DOUBLE PRECISION NOT NULL,
  rushing_open_field_yards   DOUBLE PRECISION NOT NULL,
  rushing_second_level_yards DOUBLE PRECISION NOT NULL,
  rushing_line_yards         DOUBLE PRECISION NOT NULL,

  -- RushingYardsSplit rushing_allowed
  rushing_allowed_highlight_yards    DOUBLE PRECISION NOT NULL,
  rushing_allowed_open_field_yards   DOUBLE PRECISION NOT NULL,
  rushing_allowed_second_level_yards DOUBLE PRECISION NOT NULL,
  rushing_allowed_line_yards         DOUBLE PRECISION NOT NULL,

  explosiveness              DOUBLE PRECISION NOT NULL,
  explosiveness_allowed      DOUBLE PRECISION NOT NULL,

  PRIMARY KEY (year, team_id)
);

CREATE TABLE IF NOT EXISTS cfbd.player_weighted_epa (
  year          INTEGER NOT NULL,
  athlete_id    TEXT NOT NULL,
  athlete_name  TEXT NOT NULL,
  position      TEXT NOT NULL,
  team          TEXT NOT NULL,
  conference    TEXT NOT NULL,
  wepa          DOUBLE PRECISION NOT NULL,
  plays         INTEGER NOT NULL,
  PRIMARY KEY (year, athlete_id)
);

CREATE TABLE IF NOT EXISTS cfbd.kicker_paar (
  year          INTEGER NOT NULL,
  athlete_id    TEXT NOT NULL,
  athlete_name  TEXT NOT NULL,
  team          TEXT NOT NULL,
  conference    TEXT NOT NULL,
  paar          DOUBLE PRECISION NOT NULL,
  attempts      INTEGER NOT NULL,
  PRIMARY KEY (year, athlete_id)
);

CREATE TABLE IF NOT EXISTS cfbd.team_ats (
  year              INTEGER NOT NULL,
  team_id           INTEGER NOT NULL,
  team              TEXT NOT NULL,
  conference         TEXT,
  games              INTEGER,
  ats_wins           INTEGER NOT NULL,
  ats_losses         INTEGER NOT NULL,
  ats_pushes         INTEGER NOT NULL,
  avg_cover_margin   DOUBLE PRECISION,
  PRIMARY KEY (year, team_id)
);

CREATE TABLE IF NOT EXISTS cfbd.roster_players (
  id                 TEXT PRIMARY KEY,
  first_name         TEXT NOT NULL,
  last_name          TEXT NOT NULL,
  team               TEXT NOT NULL,
  height             DOUBLE PRECISION,
  weight             INTEGER,
  jersey             INTEGER,
  position           TEXT,
  home_city          TEXT,
  home_state         TEXT,
  home_country       TEXT,
  home_latitude      DOUBLE PRECISION,
  home_longitude     DOUBLE PRECISION,
  home_county_fips   TEXT,
  recruit_ids        JSONB                 -- ListValue
);

CREATE TABLE IF NOT EXISTS cfbd.team_talent (
  year     INTEGER NOT NULL,
  team     TEXT NOT NULL,
  talent   DOUBLE PRECISION NOT NULL,
  PRIMARY KEY (year, team)
);

CREATE TABLE IF NOT EXISTS cfbd.player_stats (
  season      INTEGER NOT NULL,
  player_id   TEXT NOT NULL,
  player      TEXT NOT NULL,
  position    TEXT NOT NULL,
  team        TEXT NOT NULL,
  conference  TEXT NOT NULL,
  category    TEXT NOT NULL,
  stat_type   TEXT NOT NULL,
  stat        TEXT NOT NULL,
  PRIMARY KEY (season, player_id, category, stat_type, stat)
);

CREATE TABLE IF NOT EXISTS cfbd.team_stats (
  season      INTEGER NOT NULL,
  team        TEXT NOT NULL,
  conference  TEXT NOT NULL,
  stat_name   TEXT NOT NULL,
  stat_value  JSONB NOT NULL,   -- google.protobuf.Value
  PRIMARY KEY (season, team, stat_name)
);

-- =====================================================================
-- Matchups
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.matchups (
  matchup_id   BIGSERIAL PRIMARY KEY,
  team1        TEXT NOT NULL,
  team2        TEXT NOT NULL,
  start_year   INTEGER,
  end_year     INTEGER,
  team1_wins   INTEGER NOT NULL,
  team2_wins   INTEGER NOT NULL,
  ties         INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS cfbd.matchup_games (
  matchup_game_id BIGSERIAL PRIMARY KEY,
  matchup_id      BIGINT NOT NULL REFERENCES cfbd.matchups(matchup_id) ON DELETE CASCADE,
  season          INTEGER NOT NULL,
  week            INTEGER NOT NULL,
  season_type     TEXT NOT NULL,
  date            TEXT NOT NULL,          -- proto uses string
  neutral_site    BOOLEAN NOT NULL,
  venue           TEXT,
  home_team       TEXT NOT NULL,
  home_score      INTEGER,
  away_team       TEXT NOT NULL,
  away_score      INTEGER,
  winner          TEXT
);

CREATE INDEX IF NOT EXISTS matchup_games_matchup_id_idx
  ON cfbd.matchup_games(matchup_id);

-- =====================================================================
-- SP / SRS / Elo / FPI (jsonb for nested sub-messages)
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.team_sp (
  year                 INTEGER NOT NULL,
  team                 TEXT NOT NULL,
  conference           TEXT,

  rating               DOUBLE PRECISION,
  ranking              INTEGER,

  second_order_wins    DOUBLE PRECISION,
  sos                  DOUBLE PRECISION,

  offense              JSONB NOT NULL,  -- SpTeamOffense
  defense              JSONB NOT NULL,  -- SpTeamDefense
  special_teams        JSONB NOT NULL,  -- SpSpecialTeams

  PRIMARY KEY (year, team)
);

CREATE TABLE IF NOT EXISTS cfbd.conference_sp (
  year               INTEGER NOT NULL,
  conference         TEXT NOT NULL,
  rating             DOUBLE PRECISION NOT NULL,
  second_order_wins  DOUBLE PRECISION NOT NULL,
  sos                DOUBLE PRECISION,
  offense            JSONB NOT NULL,  -- ConferenceSpOffense
  defense            JSONB NOT NULL,  -- ConferenceSpDefense
  special_teams      JSONB NOT NULL,  -- SpSpecialTeams
  PRIMARY KEY (year, conference)
);

CREATE TABLE IF NOT EXISTS cfbd.team_srs (
  year        INTEGER NOT NULL,
  team        TEXT NOT NULL,
  conference  TEXT,
  division    TEXT,
  rating      DOUBLE PRECISION NOT NULL,
  ranking     INTEGER,
  PRIMARY KEY (year, team)
);

CREATE TABLE IF NOT EXISTS cfbd.team_elo (
  year        INTEGER NOT NULL,
  team        TEXT NOT NULL,
  conference  TEXT,
  elo         INTEGER,
  PRIMARY KEY (year, team)
);

CREATE TABLE IF NOT EXISTS cfbd.team_fpi (
  year          INTEGER NOT NULL,
  team          TEXT NOT NULL,
  conference    TEXT,
  fpi           DOUBLE PRECISION,
  resume_ranks  JSONB NOT NULL,  -- FpiResumeRanks
  efficiencies  JSONB NOT NULL,  -- FpiEfficiencies
  PRIMARY KEY (year, team)
);

-- =====================================================================
-- Polls
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.poll_weeks (
  poll_week_id BIGSERIAL PRIMARY KEY,
  season       INTEGER NOT NULL,
  season_type  TEXT NOT NULL,
  week         INTEGER NOT NULL,
  UNIQUE (season, season_type, week)
);

CREATE TABLE IF NOT EXISTS cfbd.polls (
  poll_id      BIGSERIAL PRIMARY KEY,
  poll_week_id BIGINT NOT NULL REFERENCES cfbd.poll_weeks(poll_week_id) ON DELETE CASCADE,
  poll_name    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cfbd.poll_ranks (
  poll_rank_id        BIGSERIAL PRIMARY KEY,
  poll_id             BIGINT NOT NULL REFERENCES cfbd.polls(poll_id) ON DELETE CASCADE,
  rank                INTEGER,
  team_id             INTEGER,
  school              TEXT NOT NULL,
  conference          TEXT,
  first_place_votes   INTEGER,
  points              INTEGER
);

CREATE INDEX IF NOT EXISTS polls_week_idx ON cfbd.polls(poll_week_id);
CREATE INDEX IF NOT EXISTS poll_ranks_poll_idx ON cfbd.poll_ranks(poll_id);

-- =====================================================================
-- Plays & play stats
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.plays (
  id                 TEXT PRIMARY KEY,
  drive_id           TEXT NOT NULL,
  game_id            INTEGER NOT NULL,
  drive_number       INTEGER,
  play_number        INTEGER,
  offense            TEXT NOT NULL,
  offense_conference TEXT,
  offense_score      INTEGER NOT NULL,
  defense            TEXT NOT NULL,
  home               TEXT NOT NULL,
  away               TEXT NOT NULL,
  defense_conference TEXT,
  defense_score      INTEGER NOT NULL,
  period             INTEGER NOT NULL,

  -- ClockInt32
  clock_seconds      INTEGER,
  clock_minutes      INTEGER,

  offense_timeouts   INTEGER,
  defense_timeouts   INTEGER,
  yardline           INTEGER NOT NULL,
  yards_to_goal      INTEGER NOT NULL,
  down               INTEGER NOT NULL,
  distance           INTEGER NOT NULL,
  yards_gained       INTEGER NOT NULL,
  scoring            BOOLEAN NOT NULL,
  play_type          TEXT NOT NULL,
  play_text          TEXT,
  ppa                DOUBLE PRECISION,
  wallclock          TEXT
);

CREATE INDEX IF NOT EXISTS plays_game_id_idx ON cfbd.plays(game_id);
CREATE INDEX IF NOT EXISTS plays_drive_id_idx ON cfbd.plays(drive_id);

CREATE TABLE IF NOT EXISTS cfbd.play_types (
  id            INTEGER PRIMARY KEY,
  text          TEXT NOT NULL,
  abbreviation  TEXT
);

CREATE TABLE IF NOT EXISTS cfbd.play_stat_types (
  id    INTEGER PRIMARY KEY,
  name  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cfbd.play_stats (
  play_stat_id     BIGSERIAL PRIMARY KEY,

  game_id          DOUBLE PRECISION NOT NULL,   -- proto uses double
  season           DOUBLE PRECISION NOT NULL,
  week             DOUBLE PRECISION NOT NULL,
  team             TEXT NOT NULL,
  conference       TEXT NOT NULL,
  opponent         TEXT NOT NULL,
  team_score       DOUBLE PRECISION NOT NULL,
  opponent_score   DOUBLE PRECISION NOT NULL,
  drive_id         TEXT NOT NULL,
  play_id          TEXT NOT NULL,

  period           DOUBLE PRECISION NOT NULL,

  -- ClockDouble
  clock_seconds    DOUBLE PRECISION,
  clock_minutes    DOUBLE PRECISION,

  yards_to_goal    DOUBLE PRECISION NOT NULL,
  down             DOUBLE PRECISION NOT NULL,
  distance         DOUBLE PRECISION NOT NULL,
  athlete_id       TEXT NOT NULL,
  athlete_name     TEXT NOT NULL,
  stat_type        TEXT NOT NULL,
  stat             DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS play_stats_play_id_idx ON cfbd.play_stats(play_id);
CREATE INDEX IF NOT EXISTS play_stats_game_id_idx ON cfbd.play_stats(game_id);

CREATE TABLE IF NOT EXISTS cfbd.player_search_results (
  id                   TEXT PRIMARY KEY,
  team                 TEXT NOT NULL,
  name                 TEXT NOT NULL,
  first_name           TEXT,
  last_name            TEXT,
  weight               INTEGER,
  height               DOUBLE PRECISION,
  jersey               INTEGER,
  position             TEXT NOT NULL,
  hometown             TEXT NOT NULL,
  team_color           TEXT NOT NULL,
  team_color_secondary TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cfbd.player_usage (
  season      INTEGER NOT NULL,
  id          TEXT NOT NULL,
  name        TEXT NOT NULL,
  position    TEXT NOT NULL,
  team        TEXT NOT NULL,
  conference  TEXT NOT NULL,
  usage       JSONB NOT NULL,   -- PlayerUsageSplits
  PRIMARY KEY (season, id)
);

-- =====================================================================
-- Returning production & transfers
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.returning_production (
  season                  INTEGER NOT NULL,
  team                    TEXT NOT NULL,
  conference              TEXT NOT NULL,

  total_ppa               DOUBLE PRECISION NOT NULL,
  total_passing_ppa       DOUBLE PRECISION NOT NULL,
  total_receiving_ppa     DOUBLE PRECISION NOT NULL,
  total_rushing_ppa       DOUBLE PRECISION NOT NULL,

  percent_ppa             DOUBLE PRECISION NOT NULL,
  percent_passing_ppa     DOUBLE PRECISION NOT NULL,
  percent_receiving_ppa   DOUBLE PRECISION NOT NULL,
  percent_rushing_ppa     DOUBLE PRECISION NOT NULL,

  usage                   DOUBLE PRECISION NOT NULL,
  passing_usage           DOUBLE PRECISION NOT NULL,
  receiving_usage         DOUBLE PRECISION NOT NULL,
  rushing_usage           DOUBLE PRECISION NOT NULL,

  PRIMARY KEY (season, team)
);

CREATE TABLE IF NOT EXISTS cfbd.player_transfers (
  transfer_id    BIGSERIAL PRIMARY KEY,
  season         INTEGER NOT NULL,
  first_name     TEXT NOT NULL,
  last_name      TEXT NOT NULL,
  position       TEXT NOT NULL,
  origin         TEXT NOT NULL,
  destination    TEXT,
  transfer_date  TIMESTAMPTZ NOT NULL,
  rating         DOUBLE PRECISION,
  stars          INTEGER,
  eligibility    TEXT
);

CREATE INDEX IF NOT EXISTS player_transfers_season_idx ON cfbd.player_transfers(season);

-- =====================================================================
-- Predicted points & PPA added
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.predicted_points_values (
  yard_line         INTEGER PRIMARY KEY,
  predicted_points  DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS cfbd.team_season_predicted_points_added (
  season      INTEGER NOT NULL,
  conference  TEXT NOT NULL,
  team        TEXT NOT NULL,
  offense     JSONB NOT NULL, -- TeamSeasonPredictedPointsAddedUnit
  defense     JSONB NOT NULL, -- TeamSeasonPredictedPointsAddedUnit
  PRIMARY KEY (season, team)
);

CREATE TABLE IF NOT EXISTS cfbd.team_game_predicted_points_added (
  game_id      INTEGER NOT NULL,
  season       INTEGER NOT NULL,
  week         INTEGER NOT NULL,
  season_type  TEXT NOT NULL,
  team         TEXT NOT NULL,
  conference   TEXT NOT NULL,
  opponent     TEXT NOT NULL,
  offense      JSONB NOT NULL, -- PredictedPointsAddedTotalsForGames
  defense      JSONB NOT NULL, -- PredictedPointsAddedTotalsForGames
  PRIMARY KEY (game_id, team)
);

CREATE TABLE IF NOT EXISTS cfbd.player_game_predicted_points_added (
  season      INTEGER NOT NULL,
  week        INTEGER NOT NULL,
  season_type TEXT NOT NULL,
  id          TEXT NOT NULL,
  name        TEXT NOT NULL,
  position    TEXT NOT NULL,
  team        TEXT NOT NULL,
  opponent    TEXT NOT NULL,
  average_ppa JSONB NOT NULL, -- AveragePpa
  PRIMARY KEY (season, week, season_type, id, team)
);

CREATE TABLE IF NOT EXISTS cfbd.player_season_predicted_points_added (
  season      INTEGER NOT NULL,
  id          TEXT NOT NULL,
  name        TEXT NOT NULL,
  position    TEXT NOT NULL,
  team        TEXT NOT NULL,
  conference  TEXT NOT NULL,
  average_ppa JSONB NOT NULL, -- PlayerSeasonPpaSplits
  total_ppa   JSONB NOT NULL, -- PlayerSeasonPpaSplits
  PRIMARY KEY (season, id)
);

-- =====================================================================
-- Win probability
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.play_win_probability (
  game_id                INTEGER NOT NULL,
  play_id                TEXT NOT NULL,
  play_text              TEXT NOT NULL,
  home_id                INTEGER NOT NULL,
  home                   TEXT NOT NULL,
  away_id                INTEGER NOT NULL,
  away                   TEXT NOT NULL,
  spread                 DOUBLE PRECISION NOT NULL,
  home_ball              BOOLEAN NOT NULL,
  home_score             INTEGER NOT NULL,
  away_score             INTEGER NOT NULL,
  yard_line              INTEGER NOT NULL,
  down                   INTEGER NOT NULL,
  distance               INTEGER NOT NULL,
  home_win_probability   DOUBLE PRECISION NOT NULL,
  play_number            INTEGER NOT NULL,
  PRIMARY KEY (game_id, play_id)
);

CREATE TABLE IF NOT EXISTS cfbd.pregame_win_probability (
  season               INTEGER NOT NULL,
  season_type          TEXT NOT NULL,
  week                 INTEGER NOT NULL,
  game_id              INTEGER NOT NULL,
  home_team            TEXT NOT NULL,
  away_team            TEXT NOT NULL,
  spread               DOUBLE PRECISION NOT NULL,
  home_win_probability DOUBLE PRECISION NOT NULL,
  PRIMARY KEY (game_id)
);

CREATE TABLE IF NOT EXISTS cfbd.field_goal_ep (
  yards_to_goal   INTEGER NOT NULL,
  distance        INTEGER NOT NULL,
  expected_points DOUBLE PRECISION NOT NULL,
  PRIMARY KEY (yards_to_goal, distance)
);

-- =====================================================================
-- Live game (nested -> tables)
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.live_games (
  id            INTEGER PRIMARY KEY,
  status        TEXT NOT NULL,
  period        INTEGER,
  clock         TEXT NOT NULL,
  possession    TEXT NOT NULL,
  down          INTEGER,
  distance      INTEGER,
  yards_to_goal INTEGER
);

CREATE TABLE IF NOT EXISTS cfbd.live_game_teams (
  live_game_team_id BIGSERIAL PRIMARY KEY,
  live_game_id      INTEGER NOT NULL REFERENCES cfbd.live_games(id) ON DELETE CASCADE,
  team_id           INTEGER NOT NULL,
  team              TEXT NOT NULL,
  home_away          TEXT NOT NULL,
  line_scores        INTEGER[] NOT NULL,
  points            INTEGER NOT NULL,
  drives            INTEGER NOT NULL,
  scoring_opportunities INTEGER NOT NULL,
  points_per_opportunity DOUBLE PRECISION NOT NULL,
  average_start_yard_line DOUBLE PRECISION,
  plays             INTEGER NOT NULL,
  line_yards         DOUBLE PRECISION NOT NULL,
  line_yards_per_rush DOUBLE PRECISION NOT NULL,
  second_level_yards  DOUBLE PRECISION NOT NULL,
  second_level_yards_per_rush DOUBLE PRECISION NOT NULL,
  open_field_yards    DOUBLE PRECISION NOT NULL,
  open_field_yards_per_rush DOUBLE PRECISION NOT NULL,
  epa_per_play        DOUBLE PRECISION NOT NULL,
  total_epa           DOUBLE PRECISION NOT NULL,
  passing_epa         DOUBLE PRECISION NOT NULL,
  epa_per_pass        DOUBLE PRECISION NOT NULL,
  rushing_epa         DOUBLE PRECISION NOT NULL,
  epa_per_rush        DOUBLE PRECISION NOT NULL,
  success_rate        DOUBLE PRECISION NOT NULL,
  standard_down_success_rate DOUBLE PRECISION NOT NULL,
  passing_down_success_rate  DOUBLE PRECISION NOT NULL,
  explosiveness       DOUBLE PRECISION NOT NULL,
  deserve_to_win      DOUBLE PRECISION
);

CREATE INDEX IF NOT EXISTS live_game_teams_game_idx ON cfbd.live_game_teams(live_game_id);

CREATE TABLE IF NOT EXISTS cfbd.live_game_drives (
  id               TEXT PRIMARY KEY,
  live_game_id     INTEGER NOT NULL REFERENCES cfbd.live_games(id) ON DELETE CASCADE,
  offense_id       INTEGER NOT NULL,
  offense          TEXT NOT NULL,
  defense_id       INTEGER NOT NULL,
  defense          TEXT NOT NULL,
  play_count       INTEGER NOT NULL,
  yards            INTEGER NOT NULL,
  start_period     INTEGER NOT NULL,
  start_clock      TEXT,
  start_yards_to_goal INTEGER NOT NULL,
  end_period       INTEGER,
  end_clock        TEXT,
  end_yards_to_goal INTEGER,
  duration         TEXT,
  scoring_opportunity BOOLEAN NOT NULL,
  result           TEXT NOT NULL,
  points_gained    INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS live_game_drives_game_idx ON cfbd.live_game_drives(live_game_id);

CREATE TABLE IF NOT EXISTS cfbd.live_game_plays (
  id              TEXT PRIMARY KEY,
  drive_id        TEXT NOT NULL REFERENCES cfbd.live_game_drives(id) ON DELETE CASCADE,
  home_score      INTEGER NOT NULL,
  away_score      INTEGER NOT NULL,
  period          INTEGER NOT NULL,
  clock           TEXT NOT NULL,
  wall_clock      TIMESTAMPTZ NOT NULL,
  team_id         INTEGER NOT NULL,
  team            TEXT NOT NULL,
  down            INTEGER NOT NULL,
  distance        INTEGER NOT NULL,
  yards_to_goal   INTEGER NOT NULL,
  yards_gained    INTEGER NOT NULL,
  play_type_id    INTEGER NOT NULL,
  play_type       TEXT NOT NULL,
  epa             DOUBLE PRECISION,
  garbage_time    BOOLEAN NOT NULL,
  success         BOOLEAN NOT NULL,
  rush_pass       TEXT NOT NULL,
  down_type       TEXT NOT NULL,
  play_text       TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS live_game_plays_drive_idx ON cfbd.live_game_plays(drive_id);

-- =====================================================================
-- Betting / lines / games
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.betting_games (
  id                  INTEGER PRIMARY KEY,
  season              INTEGER NOT NULL,
  season_type         TEXT NOT NULL,
  week                INTEGER NOT NULL,
  start_date          TIMESTAMPTZ NOT NULL,
  home_team_id        INTEGER NOT NULL,
  home_team           TEXT NOT NULL,
  home_conference     TEXT,
  home_classification TEXT,
  home_score          INTEGER,
  away_team_id        INTEGER NOT NULL,
  away_team           TEXT NOT NULL,
  away_conference     TEXT,
  away_classification TEXT,
  away_score          INTEGER
);

CREATE TABLE IF NOT EXISTS cfbd.game_lines (
  game_line_id     BIGSERIAL PRIMARY KEY,
  betting_game_id  INTEGER NOT NULL REFERENCES cfbd.betting_games(id) ON DELETE CASCADE,
  provider         TEXT NOT NULL,
  spread           DOUBLE PRECISION,
  formatted_spread TEXT,
  spread_open      DOUBLE PRECISION,
  over_under       DOUBLE PRECISION,
  over_under_open  DOUBLE PRECISION,
  home_moneyline   DOUBLE PRECISION,
  away_moneyline   DOUBLE PRECISION
);

CREATE INDEX IF NOT EXISTS game_lines_betting_game_idx ON cfbd.game_lines(betting_game_id);

CREATE TABLE IF NOT EXISTS cfbd.user_info (
  patron_level    DOUBLE PRECISION NOT NULL,
  remaining_calls DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS cfbd.games (
  id                           INTEGER PRIMARY KEY,
  season                       INTEGER NOT NULL,
  week                         INTEGER NOT NULL,
  season_type                  TEXT NOT NULL,
  start_date                   TIMESTAMPTZ NOT NULL,
  start_time_tbd               BOOLEAN NOT NULL,
  completed                    BOOLEAN NOT NULL,
  neutral_site                 BOOLEAN NOT NULL,
  conference_game              BOOLEAN NOT NULL,
  attendance                   INTEGER,
  venue_id                     INTEGER,
  venue                        TEXT,
  home_id                      INTEGER,
  home_team                    TEXT NOT NULL,
  home_conference              TEXT,
  home_classification          TEXT,
  home_points                  INTEGER,
  home_line_scores             JSONB,
  home_postgame_win_probability DOUBLE PRECISION,
  home_pregame_elo             INTEGER,
  home_postgame_elo            INTEGER,
  away_id                      INTEGER,
  away_team                    TEXT NOT NULL,
  away_conference              TEXT,
  away_classification          TEXT,
  away_points                  INTEGER,
  away_line_scores             JSONB,
  away_postgame_win_probability DOUBLE PRECISION,
  away_pregame_elo             INTEGER,
  away_postgame_elo            INTEGER,
  excitement_index             DOUBLE PRECISION,
  highlights                   TEXT,
  notes                        TEXT
);

-- =====================================================================
-- Game team stats
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.game_team_stats (
  id INTEGER PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS cfbd.game_team_stats_teams (
  game_team_stats_team_id BIGSERIAL PRIMARY KEY,
  game_id                 INTEGER NOT NULL REFERENCES cfbd.game_team_stats(id) ON DELETE CASCADE,
  team_id                 INTEGER NOT NULL,
  team                    TEXT NOT NULL,
  conference              TEXT,
  home_away               TEXT NOT NULL,
  points                  INTEGER
);

CREATE TABLE IF NOT EXISTS cfbd.game_team_stats_team_stats (
  game_team_stats_team_stat_id BIGSERIAL PRIMARY KEY,
  game_team_stats_team_id      BIGINT NOT NULL REFERENCES cfbd.game_team_stats_teams(game_team_stats_team_id) ON DELETE CASCADE,
  category                     TEXT NOT NULL,
  stat                         TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS game_team_stats_teams_game_idx ON cfbd.game_team_stats_teams(game_id);

-- =====================================================================
-- Game player stats (deep nesting; keep as JSONB unless you want this normalized too)
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.game_player_stats (
  id     INTEGER PRIMARY KEY,
  teams  JSONB NOT NULL   -- repeated GamePlayerStatsTeam with categories/types/athletes
);

-- =====================================================================
-- Media & weather
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.game_media (
  id            INTEGER PRIMARY KEY,
  season        INTEGER NOT NULL,
  week          INTEGER NOT NULL,
  season_type   TEXT NOT NULL,
  start_time    TIMESTAMPTZ NOT NULL,
  is_start_time_tbd BOOLEAN NOT NULL,
  home_team     TEXT NOT NULL,
  home_conference TEXT,
  away_team     TEXT NOT NULL,
  away_conference TEXT,
  media_type    TEXT NOT NULL,
  outlet        TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cfbd.game_weather (
  id            INTEGER PRIMARY KEY,
  season        INTEGER NOT NULL,
  week          INTEGER NOT NULL,
  season_type   TEXT NOT NULL,
  start_time    TIMESTAMPTZ NOT NULL,
  game_indoors  BOOLEAN NOT NULL,
  home_team     TEXT NOT NULL,
  home_conference TEXT,
  away_team     TEXT NOT NULL,
  away_conference TEXT,
  venue_id      INTEGER,
  venue         TEXT,
  temperature   DOUBLE PRECISION,
  dew_point     DOUBLE PRECISION,
  humidity      DOUBLE PRECISION,
  precipitation DOUBLE PRECISION,
  snowfall      DOUBLE PRECISION,
  wind_direction DOUBLE PRECISION,
  wind_speed    DOUBLE PRECISION,
  pressure      DOUBLE PRECISION,
  weather_condition_code DOUBLE PRECISION,
  weather_condition TEXT
);

-- =====================================================================
-- Records / calendar / scoreboard
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.team_records (
  year            INTEGER NOT NULL,
  team_id         INTEGER,
  team            TEXT NOT NULL,
  classification  TEXT,
  conference      TEXT NOT NULL,
  division        TEXT NOT NULL,
  expected_wins   DOUBLE PRECISION,

  total                 JSONB NOT NULL, -- TeamRecord
  conference_games      JSONB NOT NULL,
  home_games            JSONB NOT NULL,
  away_games            JSONB NOT NULL,
  neutral_site_games    JSONB NOT NULL,
  regular_season        JSONB NOT NULL,
  postseason            JSONB NOT NULL,

  PRIMARY KEY (year, team)
);

CREATE TABLE IF NOT EXISTS cfbd.calendar_weeks (
  season          INTEGER NOT NULL,
  week            INTEGER NOT NULL,
  season_type     TEXT NOT NULL,
  start_date      TIMESTAMPTZ NOT NULL,
  end_date        TIMESTAMPTZ NOT NULL,
  first_game_start TIMESTAMPTZ, -- deprecated
  last_game_start  TIMESTAMPTZ, -- deprecated
  PRIMARY KEY (season, season_type, week)
);

CREATE TABLE IF NOT EXISTS cfbd.scoreboards (
  id              INTEGER PRIMARY KEY,
  start_date      TIMESTAMPTZ NOT NULL,
  start_time_tbd  BOOLEAN NOT NULL,
  tv              TEXT,
  neutral_site    BOOLEAN NOT NULL,
  conference_game BOOLEAN NOT NULL,
  status          TEXT NOT NULL,
  period          INTEGER,
  clock           TEXT,
  situation       TEXT,
  possession      TEXT,
  last_play       TEXT,
  venue           JSONB NOT NULL,   -- google.protobuf.Struct
  home_team       JSONB NOT NULL,   -- google.protobuf.Struct
  away_team       JSONB NOT NULL,   -- google.protobuf.Struct
  weather         JSONB NOT NULL,   -- google.protobuf.Struct
  betting         JSONB NOT NULL    -- google.protobuf.Struct
);

-- =====================================================================
-- Drives
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.drives (
  id                    TEXT PRIMARY KEY,
  game_id               INTEGER NOT NULL,
  offense               TEXT NOT NULL,
  offense_conference    TEXT,
  defense               TEXT NOT NULL,
  defense_conference    TEXT,
  drive_number          INTEGER,
  scoring               BOOLEAN NOT NULL,
  start_period          INTEGER NOT NULL,
  start_yardline        INTEGER NOT NULL,
  start_yards_to_goal   INTEGER NOT NULL,
  start_time_seconds    INTEGER,
  start_time_minutes    INTEGER,
  end_period            INTEGER NOT NULL,
  end_yardline          INTEGER NOT NULL,
  end_yards_to_goal     INTEGER NOT NULL,
  end_time_seconds      INTEGER,
  end_time_minutes      INTEGER,
  elapsed_seconds       INTEGER,
  elapsed_minutes       INTEGER,
  plays                 INTEGER NOT NULL,
  yards                 INTEGER NOT NULL,
  drive_result          TEXT NOT NULL,
  is_home_offense       BOOLEAN NOT NULL,
  start_offense_score   INTEGER NOT NULL,
  start_defense_score   INTEGER NOT NULL,
  end_offense_score     INTEGER NOT NULL,
  end_defense_score     INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS drives_game_id_idx ON cfbd.drives(game_id);

-- =====================================================================
-- Draft
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.draft_picks (
  draft_pick_id                 BIGSERIAL PRIMARY KEY,
  college_athlete_id            INTEGER,
  nfl_athlete_id                INTEGER,
  college_id                    INTEGER NOT NULL,
  college_team                  TEXT NOT NULL,
  college_conference            TEXT,
  nfl_team_id                   INTEGER NOT NULL,
  nfl_team                      TEXT NOT NULL,
  year                          INTEGER NOT NULL,
  overall                       INTEGER NOT NULL,
  round                         INTEGER NOT NULL,
  pick                          INTEGER NOT NULL,
  name                          TEXT NOT NULL,
  position                      TEXT NOT NULL,
  height                        DOUBLE PRECISION,
  weight                        INTEGER,
  pre_draft_ranking             INTEGER,
  pre_draft_position_ranking    INTEGER,
  pre_draft_grade               INTEGER,

  -- DraftPickHometownInfo (all strings in proto)
  hometown_county_fips          TEXT,
  hometown_longitude            TEXT,
  hometown_latitude             TEXT,
  hometown_country              TEXT,
  hometown_state                TEXT,
  hometown_city                 TEXT
);

CREATE INDEX IF NOT EXISTS draft_picks_year_idx ON cfbd.draft_picks(year);

-- =====================================================================
-- Coaches
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.coaches (
  coach_id     BIGSERIAL PRIMARY KEY,
  first_name   TEXT NOT NULL,
  last_name    TEXT NOT NULL,
  hire_date    TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS cfbd.coach_seasons (
  coach_season_id BIGSERIAL PRIMARY KEY,
  coach_id        BIGINT NOT NULL REFERENCES cfbd.coaches(coach_id) ON DELETE CASCADE,
  school          TEXT NOT NULL,
  year            INTEGER NOT NULL,
  games           INTEGER NOT NULL,
  wins            INTEGER NOT NULL,
  losses          INTEGER NOT NULL,
  ties            INTEGER NOT NULL,
  preseason_rank  INTEGER,
  postseason_rank INTEGER,
  srs             DOUBLE PRECISION,
  sp_overall      DOUBLE PRECISION,
  sp_offense      DOUBLE PRECISION,
  sp_defense      DOUBLE PRECISION,
  UNIQUE (coach_id, school, year)
);

-- =====================================================================
-- Recruiting
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.recruits (
  id               TEXT PRIMARY KEY,
  athlete_id       TEXT,
  recruit_type     TEXT NOT NULL,
  year             INTEGER NOT NULL,
  ranking          INTEGER,
  name             TEXT NOT NULL,
  school           TEXT,
  committed_to     TEXT,
  position         TEXT,
  height           DOUBLE PRECISION,
  weight           INTEGER,
  stars            INTEGER NOT NULL,
  rating           DOUBLE PRECISION NOT NULL,
  city             TEXT,
  state_province   TEXT,
  country          TEXT,

  hometown_fips_code  TEXT,
  hometown_longitude  DOUBLE PRECISION,
  hometown_latitude   DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS cfbd.team_recruiting_rankings (
  year     INTEGER NOT NULL,
  rank     INTEGER NOT NULL,
  team     TEXT NOT NULL,
  points   DOUBLE PRECISION NOT NULL,
  PRIMARY KEY (year, team)
);

CREATE TABLE IF NOT EXISTS cfbd.aggregated_team_recruiting (
  team            TEXT NOT NULL,
  conference      TEXT NOT NULL,
  position_group  TEXT,
  average_rating  DOUBLE PRECISION NOT NULL,
  total_rating    DOUBLE PRECISION NOT NULL,
  commits         INTEGER NOT NULL,
  average_stars   DOUBLE PRECISION NOT NULL,
  PRIMARY KEY (team, conference, COALESCE(position_group,''))
);

-- =====================================================================
-- Game havoc stats (kept as jsonb for the side message)
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.game_havoc_stats (
  game_id              INTEGER NOT NULL,
  season               INTEGER NOT NULL,
  season_type          TEXT NOT NULL,
  week                 INTEGER NOT NULL,
  team                 TEXT NOT NULL,
  conference            TEXT,
  opponent             TEXT NOT NULL,
  opponent_conference   TEXT,
  offense              JSONB NOT NULL,   -- GameHavocStatSide
  defense              JSONB NOT NULL,   -- GameHavocStatSide
  PRIMARY KEY (game_id, team)
);

-- =====================================================================
-- Advanced Season Stats (FULLY NORMALIZED)
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.adv_rate_metrics (
  adv_rate_metrics_id BIGSERIAL PRIMARY KEY,
  explosiveness DOUBLE PRECISION,
  success_rate  DOUBLE PRECISION,
  total_ppa     DOUBLE PRECISION,
  ppa           DOUBLE PRECISION,
  rate          DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS cfbd.adv_havoc (
  adv_havoc_id BIGSERIAL PRIMARY KEY,
  db          DOUBLE PRECISION,
  front_seven DOUBLE PRECISION,
  total       DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS cfbd.adv_field_position (
  adv_field_position_id BIGSERIAL PRIMARY KEY,
  average_predicted_points DOUBLE PRECISION,
  average_start            DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS cfbd.adv_season_stat_side (
  adv_season_stat_side_id BIGSERIAL PRIMARY KEY,

  passing_plays_id   BIGINT REFERENCES cfbd.adv_rate_metrics(adv_rate_metrics_id) ON DELETE SET NULL,
  rushing_plays_id   BIGINT REFERENCES cfbd.adv_rate_metrics(adv_rate_metrics_id) ON DELETE SET NULL,
  passing_downs_id   BIGINT REFERENCES cfbd.adv_rate_metrics(adv_rate_metrics_id) ON DELETE SET NULL,
  standard_downs_id  BIGINT REFERENCES cfbd.adv_rate_metrics(adv_rate_metrics_id) ON DELETE SET NULL,
  havoc_id           BIGINT REFERENCES cfbd.adv_havoc(adv_havoc_id) ON DELETE SET NULL,
  field_position_id  BIGINT REFERENCES cfbd.adv_field_position(adv_field_position_id) ON DELETE SET NULL,

  points_per_opportunity DOUBLE PRECISION,
  total_opportunies      INTEGER,
  open_field_yards_total INTEGER,
  open_field_yards       DOUBLE PRECISION,
  second_level_yards_total INTEGER,
  second_level_yards       DOUBLE PRECISION,
  line_yards_total         INTEGER,
  line_yards               DOUBLE PRECISION,
  stuff_rate             DOUBLE PRECISION,
  power_success          DOUBLE PRECISION,

  explosiveness          DOUBLE PRECISION,
  success_rate           DOUBLE PRECISION,
  total_ppa              DOUBLE PRECISION,
  ppa                    DOUBLE PRECISION,

  drives                 INTEGER,
  plays                  INTEGER
);

CREATE TABLE IF NOT EXISTS cfbd.advanced_season_stats_normalized (
  season      INTEGER NOT NULL,
  team        TEXT NOT NULL,
  conference  TEXT NOT NULL,

  offense_side_id BIGINT NOT NULL REFERENCES cfbd.adv_season_stat_side(adv_season_stat_side_id) ON DELETE RESTRICT,
  defense_side_id BIGINT NOT NULL REFERENCES cfbd.adv_season_stat_side(adv_season_stat_side_id) ON DELETE RESTRICT,

  PRIMARY KEY (season, team)
);

CREATE INDEX IF NOT EXISTS advanced_season_stats_norm_offense_idx
  ON cfbd.advanced_season_stats_normalized(offense_side_id);
CREATE INDEX IF NOT EXISTS advanced_season_stats_norm_defense_idx
  ON cfbd.advanced_season_stats_normalized(defense_side_id);

-- =====================================================================
-- Advanced Game Stats (FULLY NORMALIZED)
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.adv_game_play_metrics (
  adv_game_play_metrics_id BIGSERIAL PRIMARY KEY,
  explosiveness DOUBLE PRECISION,
  success_rate  DOUBLE PRECISION,
  total_ppa     DOUBLE PRECISION,
  ppa           DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS cfbd.adv_game_down_metrics (
  adv_game_down_metrics_id BIGSERIAL PRIMARY KEY,
  explosiveness DOUBLE PRECISION,
  success_rate  DOUBLE PRECISION,
  ppa           DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS cfbd.adv_game_stat_side (
  adv_game_stat_side_id BIGSERIAL PRIMARY KEY,

  passing_plays_id   BIGINT REFERENCES cfbd.adv_game_play_metrics(adv_game_play_metrics_id) ON DELETE SET NULL,
  rushing_plays_id   BIGINT REFERENCES cfbd.adv_game_play_metrics(adv_game_play_metrics_id) ON DELETE SET NULL,
  passing_downs_id   BIGINT REFERENCES cfbd.adv_game_down_metrics(adv_game_down_metrics_id) ON DELETE SET NULL,
  standard_downs_id  BIGINT REFERENCES cfbd.adv_game_down_metrics(adv_game_down_metrics_id) ON DELETE SET NULL,

  open_field_yards_total INTEGER,
  open_field_yards       DOUBLE PRECISION,
  second_level_yards_total INTEGER,
  second_level_yards       DOUBLE PRECISION,
  line_yards_total         INTEGER,
  line_yards               DOUBLE PRECISION,

  stuff_rate        DOUBLE PRECISION,
  power_success     DOUBLE PRECISION,

  explosiveness     DOUBLE PRECISION,
  success_rate      DOUBLE PRECISION,
  total_ppa         DOUBLE PRECISION,
  ppa               DOUBLE PRECISION,

  drives            INTEGER,
  plays             INTEGER
);

CREATE TABLE IF NOT EXISTS cfbd.advanced_game_stats_normalized (
  game_id      INTEGER NOT NULL,
  season       INTEGER NOT NULL,
  season_type  TEXT NOT NULL,
  week         INTEGER NOT NULL,
  team         TEXT NOT NULL,
  opponent     TEXT NOT NULL,

  offense_side_id BIGINT NOT NULL REFERENCES cfbd.adv_game_stat_side(adv_game_stat_side_id) ON DELETE RESTRICT,
  defense_side_id BIGINT NOT NULL REFERENCES cfbd.adv_game_stat_side(adv_game_stat_side_id) ON DELETE RESTRICT,

  PRIMARY KEY (game_id, team)
);

CREATE INDEX IF NOT EXISTS advanced_game_stats_norm_offense_idx
  ON cfbd.advanced_game_stats_normalized(offense_side_id);
CREATE INDEX IF NOT EXISTS advanced_game_stats_norm_defense_idx
  ON cfbd.advanced_game_stats_normalized(defense_side_id);

-- =====================================================================
-- Advanced Box Score (FULLY NORMALIZED)
-- =====================================================================

CREATE TABLE IF NOT EXISTS cfbd.advanced_box_score_game_info (
  abs_game_info_id BIGSERIAL PRIMARY KEY,
  excitement    DOUBLE PRECISION NOT NULL,
  home_winner   BOOLEAN NOT NULL,
  away_win_prob DOUBLE PRECISION NOT NULL,
  away_points   INTEGER NOT NULL,
  away_team     TEXT NOT NULL,
  home_win_prob DOUBLE PRECISION NOT NULL,
  home_points   INTEGER NOT NULL,
  home_team     TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cfbd.advanced_box_scores (
  advanced_box_score_id BIGSERIAL PRIMARY KEY,
  game_info_id BIGINT NOT NULL REFERENCES cfbd.advanced_box_score_game_info(abs_game_info_id) ON DELETE RESTRICT
);

-- Common "by quarter" (StatsByQuarter)
CREATE TABLE IF NOT EXISTS cfbd.stats_by_quarter (
  stats_by_quarter_id BIGSERIAL PRIMARY KEY,
  total    DOUBLE PRECISION NOT NULL,
  quarter1 DOUBLE PRECISION,
  quarter2 DOUBLE PRECISION,
  quarter3 DOUBLE PRECISION,
  quarter4 DOUBLE PRECISION
);

-- TeamFieldPosition
CREATE TABLE IF NOT EXISTS cfbd.abs_team_field_position (
  abs_team_field_position_id BIGSERIAL PRIMARY KEY,
  advanced_box_score_id BIGINT NOT NULL REFERENCES cfbd.advanced_box_scores(advanced_box_score_id) ON DELETE CASCADE,
  team TEXT NOT NULL,
  average_start DOUBLE PRECISION NOT NULL,
  average_starting_predicted_points DOUBLE PRECISION NOT NULL
);

-- TeamScoringOpportunities
CREATE TABLE IF NOT EXISTS cfbd.abs_team_scoring_opportunities (
  abs_team_scoring_opportunities_id BIGSERIAL PRIMARY KEY,
  advanced_box_score_id BIGINT NOT NULL REFERENCES cfbd.advanced_box_scores(advanced_box_score_id) ON DELETE CASCADE,
  team TEXT NOT NULL,
  opportunities INTEGER NOT NULL,
  points INTEGER NOT NULL,
  points_per_opportunity DOUBLE PRECISION NOT NULL
);

-- TeamHavoc
CREATE TABLE IF NOT EXISTS cfbd.abs_team_havoc (
  abs_team_havoc_id BIGSERIAL PRIMARY KEY,
  advanced_box_score_id BIGINT NOT NULL REFERENCES cfbd.advanced_box_scores(advanced_box_score_id) ON DELETE CASCADE,
  team TEXT NOT NULL,
  total DOUBLE PRECISION NOT NULL,
  front_seven DOUBLE PRECISION NOT NULL,
  db DOUBLE PRECISION NOT NULL
);

-- TeamRushingStats
CREATE TABLE IF NOT EXISTS cfbd.abs_team_rushing_stats (
  abs_team_rushing_stats_id BIGSERIAL PRIMARY KEY,
  advanced_box_score_id BIGINT NOT NULL REFERENCES cfbd.advanced_box_scores(advanced_box_score_id) ON DELETE CASCADE,
  team TEXT NOT NULL,
  power_success DOUBLE PRECISION NOT NULL,
  stuff_rate DOUBLE PRECISION NOT NULL,
  line_yards DOUBLE PRECISION NOT NULL,
  line_yards_average DOUBLE PRECISION NOT NULL,
  second_level_yards DOUBLE PRECISION NOT NULL,
  second_level_yards_average DOUBLE PRECISION NOT NULL,
  open_field_yards DOUBLE PRECISION NOT NULL,
  open_field_yards_average DOUBLE PRECISION NOT NULL
);

-- TeamExplosiveness
CREATE TABLE IF NOT EXISTS cfbd.abs_team_explosiveness (
  abs_team_explosiveness_id BIGSERIAL PRIMARY KEY,
  advanced_box_score_id BIGINT NOT NULL REFERENCES cfbd.advanced_box_scores(advanced_box_score_id) ON DELETE CASCADE,
  team TEXT NOT NULL,
  overall_id BIGINT NOT NULL REFERENCES cfbd.stats_by_quarter(stats_by_quarter_id) ON DELETE RESTRICT
);

-- TeamSuccessRates
CREATE TABLE IF NOT EXISTS cfbd.abs_team_success_rates (
  abs_team_success_rates_id BIGSERIAL PRIMARY KEY,
  advanced_box_score_id BIGINT NOT NULL REFERENCES cfbd.advanced_box_scores(advanced_box_score_id) ON DELETE CASCADE,
  team TEXT NOT NULL,
  overall_id        BIGINT NOT NULL REFERENCES cfbd.stats_by_quarter(stats_by_quarter_id) ON DELETE RESTRICT,
  standard_downs_id BIGINT NOT NULL REFERENCES cfbd.stats_by_quarter(stats_by_quarter_id) ON DELETE RESTRICT,
  passing_downs_id  BIGINT NOT NULL REFERENCES cfbd.stats_by_quarter(stats_by_quarter_id) ON DELETE RESTRICT
);

-- TeamPPA lists: "ppa" and "cumulative_ppa"
CREATE TABLE IF NOT EXISTS cfbd.abs_team_ppa (
  abs_team_ppa_id BIGSERIAL PRIMARY KEY,
  advanced_box_score_id BIGINT NOT NULL REFERENCES cfbd.advanced_box_scores(advanced_box_score_id) ON DELETE CASCADE,
  team TEXT NOT NULL,
  plays INTEGER NOT NULL,
  overall_id BIGINT NOT NULL REFERENCES cfbd.stats_by_quarter(stats_by_quarter_id) ON DELETE RESTRICT,
  passing_id BIGINT NOT NULL REFERENCES cfbd.stats_by_quarter(stats_by_quarter_id) ON DELETE RESTRICT,
  rushing_id BIGINT NOT NULL REFERENCES cfbd.stats_by_quarter(stats_by_quarter_id) ON DELETE RESTRICT,
  kind TEXT NOT NULL CHECK (kind IN ('ppa','cumulative_ppa'))
);

-- PlayerStatsByQuarter (player shape includes rushing/passing totals too)
CREATE TABLE IF NOT EXISTS cfbd.player_stats_by_quarter (
  player_stats_by_quarter_id BIGSERIAL PRIMARY KEY,
  total    DOUBLE PRECISION NOT NULL,
  quarter1 DOUBLE PRECISION,
  quarter2 DOUBLE PRECISION,
  quarter3 DOUBLE PRECISION,
  quarter4 DOUBLE PRECISION,
  rushing  DOUBLE PRECISION NOT NULL,
  passing  DOUBLE PRECISION NOT NULL
);

-- PlayerPPA
CREATE TABLE IF NOT EXISTS cfbd.abs_player_ppa (
  abs_player_ppa_id BIGSERIAL PRIMARY KEY,
  advanced_box_score_id BIGINT NOT NULL REFERENCES cfbd.advanced_box_scores(advanced_box_score_id) ON DELETE CASCADE,
  player   TEXT NOT NULL,
  team     TEXT NOT NULL,
  position TEXT NOT NULL,
  average_id    BIGINT NOT NULL REFERENCES cfbd.player_stats_by_quarter(player_stats_by_quarter_id) ON DELETE RESTRICT,
  cumulative_id BIGINT NOT NULL REFERENCES cfbd.player_stats_by_quarter(player_stats_by_quarter_id) ON DELETE RESTRICT
);

-- PlayerGameUsage uses same quarter shape as PlayerStatsByQuarter in proto (but separate message)
CREATE TABLE IF NOT EXISTS cfbd.player_game_usage_quarters (
  player_game_usage_quarters_id BIGSERIAL PRIMARY KEY,
  total    DOUBLE PRECISION NOT NULL,
  quarter1 DOUBLE PRECISION,
  quarter2 DOUBLE PRECISION,
  quarter3 DOUBLE PRECISION,
  quarter4 DOUBLE PRECISION,
  rushing  DOUBLE PRECISION NOT NULL,
  passing  DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS cfbd.abs_player_game_usage (
  abs_player_game_usage_id BIGSERIAL PRIMARY KEY,
  advanced_box_score_id BIGINT NOT NULL REFERENCES cfbd.advanced_box_scores(advanced_box_score_id) ON DELETE CASCADE,
  player   TEXT NOT NULL,
  team     TEXT NOT NULL,
  position TEXT NOT NULL,
  usage_quarters_id BIGINT NOT NULL REFERENCES cfbd.player_game_usage_quarters(player_game_usage_quarters_id) ON DELETE RESTRICT
);

-- Indexes for box score lookups
CREATE INDEX IF NOT EXISTS abs_team_field_position_abs_idx
  ON cfbd.abs_team_field_position(advanced_box_score_id);
CREATE INDEX IF NOT EXISTS abs_team_scoring_opps_abs_idx
  ON cfbd.abs_team_scoring_opportunities(advanced_box_score_id);
CREATE INDEX IF NOT EXISTS abs_team_havoc_abs_idx
  ON cfbd.abs_team_havoc(advanced_box_score_id);
CREATE INDEX IF NOT EXISTS abs_team_rushing_abs_idx
  ON cfbd.abs_team_rushing_stats(advanced_box_score_id);
CREATE INDEX IF NOT EXISTS abs_team_expl_abs_idx
  ON cfbd.abs_team_explosiveness(advanced_box_score_id);
CREATE INDEX IF NOT EXISTS abs_team_sr_abs_idx
  ON cfbd.abs_team_success_rates(advanced_box_score_id);
CREATE INDEX IF NOT EXISTS abs_team_ppa_abs_idx
  ON cfbd.abs_team_ppa(advanced_box_score_id);
CREATE INDEX IF NOT EXISTS abs_player_ppa_abs_idx
  ON cfbd.abs_player_ppa(advanced_box_score_id);
CREATE INDEX IF NOT EXISTS abs_player_usage_abs_idx
  ON cfbd.abs_player_game_usage(advanced_box_score_id);

COMMIT;

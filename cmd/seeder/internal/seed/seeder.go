package seed

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/clintrovert/cfbd-etl/seeder/internal/db"
	"github.com/clintrovert/cfbd-go/cfbd"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

// var supportedYears = []int32{
//    2005, 2006, 2007, 2008, 2009, 2010, 2011, 2012, 2013, 2014, 2015, 2016,
//    2017, 2018, 2019, 2020, 2021, 2022, 2023, 2024, 2025,
// }

var supportedYears = []int32{2024, 2025}

type Seeder struct {
	db           *db.Database
	api          *cfbd.Client
	ctx          context.Context
	throttler    *rate.Limiter
	throttleLock sync.Mutex
}

// NewSeeder todo:describe.
func NewSeeder(
	db *db.Database,
	api *cfbd.Client,
	throttle *rate.Limiter,
) (*Seeder, error) {
	return &Seeder{
		db:        db,
		api:       api,
		throttler: throttle,
	}, nil
}

// throttle waits for the rate limiter to allow a request.
// This should be called before making any API request.
func (s *Seeder) throttle(ctx context.Context) error {
	s.throttleLock.Lock()
	throttle := s.throttler
	s.throttleLock.Unlock()

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := throttle.Wait(waitCtx); err != nil {
		return fmt.Errorf("rate limiter wait failed: %w", err)
	}

	return nil
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
	if err := s.throttle(s.ctx); err != nil {
		return fmt.Errorf("failed to wait for rate limit; %w", err)
	}

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
	if err := s.throttle(s.ctx); err != nil {
		return fmt.Errorf("failed to wait for rate limit; %w", err)
	}

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
	if err := s.throttle(s.ctx); err != nil {
		return fmt.Errorf("failed to wait for rate limit; %w", err)
	}

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
	if err := s.throttle(s.ctx); err != nil {
		return fmt.Errorf("failed to wait for rate limit; %w", err)
	}

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
	if err := s.throttle(s.ctx); err != nil {
		return fmt.Errorf("failed to wait for rate limit; %w", err)
	}

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
	if err := s.throttle(s.ctx); err != nil {
		return fmt.Errorf("failed to wait for rate limit; %w", err)
	}

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
	if err := s.throttle(s.ctx); err != nil {
		return fmt.Errorf("failed to wait for rate limit; %w", err)
	}

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
	if err := s.throttle(s.ctx); err != nil {
		return fmt.Errorf("failed to wait for rate limit; %w", err)
	}

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
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

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
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

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
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

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
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		// GetPlays requires both a year and a week to be specified.
		// We must query GetCalendar first to get the available weeks
		// for each year.
		weeks, err := s.api.GetCalendar(
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

		for _, week := range weeks {
			if err = s.throttle(s.ctx); err != nil {
				return fmt.Errorf("failed to wait for rate limit; %w", err)
			}

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
				return fmt.Errorf(
					"failed to get plays for year %d, week %d, season_type %s; %w",
					year, week.GetWeek(), week.GetSeasonType(), err,
				)
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
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		// GetPlayStats requires both a year and a week to be specified.
		// We must query GetCalendar first to get the available weeks
		// for each year.
		calendarWeeks, err := s.api.GetCalendar(
			s.ctx, cfbd.GetCalendarRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get calendar for play stats",
				"year", int32ToString(year),
				"err", err,
			)
			return fmt.Errorf("failed to get calendar for year %d; %w", year, err)
		}

		for _, week := range calendarWeeks {
			if err = s.throttle(s.ctx); err != nil {
				return fmt.Errorf("failed to wait for rate limit; %w", err)
			}

			playStats, err := s.api.GetPlayStats(s.ctx, cfbd.GetPlayStatsRequest{
				Year:       year,
				Week:       week.GetWeek(),
				SeasonType: week.GetSeasonType(),
			})
			if err != nil {
				slog.Error(
					"failed to get play stats",
					"year", int32ToString(year),
					"week", int32ToString(week.GetWeek()),
					"season_type", week.GetSeasonType(),
					"err", err,
				)
				return fmt.Errorf(
					"failed to get playstats for year %d, week %d, szntype %s; %w",
					year, week.GetWeek(), week.GetSeasonType(), err,
				)
			}

			if len(playStats) > 0 {
				if err = s.db.InsertPlayStats(s.ctx, playStats); err != nil {
					slog.Error("failed to insert play stats", "err", err)
					return fmt.Errorf("failed to insert play stats; %w", err)
				}

				totalInserted += len(playStats)
				slog.Info("inserted play stats",
					"year", int32ToString(year),
					"week", int32ToString(week.GetWeek()),
					"season_type", week.GetSeasonType(),
					"count", len(playStats),
					"total", totalInserted,
				)
			}
		}
	}

	slog.Info("play stats successfully inserted", "total_count", totalInserted)
	return nil
}

func (s *Seeder) SeedGameTeamStats() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		stats, err := s.api.GetGameTeams(
			s.ctx, cfbd.GetGameTeamsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get game team stats",
				"year", int32ToString(year),
				"err", err,
			)
			return fmt.Errorf(
				"failed to get game team stats for year %d; %w", year, err,
			)
		}

		if len(stats) > 0 {
			if err := s.db.InsertGameTeamStats(s.ctx, stats); err != nil {
				slog.Error("failed to insert game team stats", "err", err)
				return fmt.Errorf("failed to insert game team stats; %w", err)
			}
			totalInserted += len(stats)
			slog.Info("inserted game team stats",
				"year", int32ToString(year),
				"count", len(stats),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"game team stats successfully inserted",
		"total_count", totalInserted,
	)
	return nil
}

func (s *Seeder) SeedGamePlayerStats() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		stats, err := s.api.GetGamePlayers(
			s.ctx, cfbd.GetGamePlayersRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get game player stats",
				"year", int32ToString(year),
				"err", err,
			)
			return fmt.Errorf(
				"failed to get game player stats for year %d; %w", year, err,
			)
		}

		if len(stats) > 0 {
			if err := s.db.InsertGamePlayerStats(s.ctx, stats); err != nil {
				slog.Error("failed to insert game player stats", "err", err)
				return fmt.Errorf("failed to insert game player stats; %w", err)
			}
			totalInserted += len(stats)
			slog.Info("inserted game player stats",
				"year", int32ToString(year),
				"count", len(stats),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"game player stats successfully inserted", "total_count", totalInserted,
	)
	return nil
}

func (s *Seeder) SeedWinProbability() error {
	for _, year := range supportedYears {
		slog.Info("seeding win probability", "year", year)

		gameIDs, err := s.db.GetGameIDs(s.ctx, int(year))
		if err != nil {
			return fmt.Errorf("failed to get game IDs for year %d: %w", year, err)
		}

		// Process games in batches to avoid overwhelming the API
		// or process one by one if rate limit is tight.
		// Seeder has rate limiter usage in `fetch` method but getting WP is per
		// game.
		// Use a worker pool or simple loop? Simple loop with concurrency control
		// via errgroup is typical in this file.
		// However, fetching one by one for thousands of games might be slow.
		// Let's use the pattern from other functions if possible, or simple loop
		// with error group.
		// Given we have GetWinProbability for a specific game, we loop.

		// NOTE: GetWinProbability might accept multiple IDs?
		// Check cfbd_doc.txt for GetWinProbabilityRequest.
		// Step 447 output: type GetWinProbabilityRequest struct { GameId int32 ...}
		// It creates a query param. Usually CFBD allows filtering by year/team OR
		// specific game ID.
		// If it allows filtering by year, we can do bulk fetch!
		// Let's check if GetWinProbabilityRequest has Year field.
		// Step 447 didn't show fields inside.
		// Let's assume we iterate if we can't bulk.

		// Actually, let's verify if GetWinProbability supports 'Year'.
		// If it does, we don't need game IDs.
		// I will check `cfbd_doc.txt` again for Request struct fields.
		// If not, I follow the plan of iterating IDs.

		// To be safe and quick, I'll write the iteration logic assuming per-game
		// fetch for now, but check filtering support first.

		group, ctx := errgroup.WithContext(s.ctx)
		group.SetLimit(10) // Limit concurrency

		for _, gameID := range gameIDs {
			gid := gameID
			group.Go(func() error {
				if err := s.throttle(ctx); err != nil {
					return err
				}
				plays, err := s.api.GetWinProbability(
					ctx, cfbd.GetWinProbabilityRequest{GameID: gid},
				)
				if err != nil {
					slog.Warn(
						"failed to get win probability",
						"year", year,
						"game_id", gid,
						"err", err,
					)
					return nil // Continue despite error
				}

				if len(plays) == 0 {
					return nil
				}

				return s.db.InsertPlayWinProbability(ctx, plays)
			})
		}

		if err := group.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedAdvancedBoxScore() error {
	for _, year := range supportedYears {
		slog.Info("seeding advanced box scores", "year", year)

		gameIDs, err := s.db.GetGameIDs(s.ctx, int(year))
		if err != nil {
			return fmt.Errorf("failed to get game IDs for year %d: %w", year, err)
		}

		// Batch inserts for box scores
		var mu sync.Mutex
		batch := make(map[int32]*cfbd.AdvancedBoxScore)

		group, ctx := errgroup.WithContext(s.ctx)
		group.SetLimit(10)

		for _, gameID := range gameIDs {
			gid := gameID
			group.Go(func() error {
				if err := s.throttle(ctx); err != nil {
					return err
				}
				score, err := s.api.GetAdvancedBoxScore(
					ctx, cfbd.GetAdvancedBoxScoreRequest{GameID: gid},
				)
				if err != nil {
					slog.Warn(
						"failed to get advanced box score",
						"year", year, "game_id", gid, "err", err,
					)
					return nil
				}

				mu.Lock()
				batch[gid] = score
				if len(batch) >= 100 {
					// Flush batch
					params := batch
					batch = make(map[int32]*cfbd.AdvancedBoxScore)
					mu.Unlock()
					return s.db.InsertAdvancedBoxScores(ctx, params)
				}
				mu.Unlock()
				return nil
			})
		}

		if err := group.Wait(); err != nil {
			return err
		}

		// Flush remaining
		if len(batch) > 0 {
			if err := s.db.InsertAdvancedBoxScores(s.ctx, batch); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Seeder) SeedGameWeather() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		weather, err := s.api.GetGameWeather(
			s.ctx, cfbd.GetGameWeatherRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get game weather",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf("failed to get game weather for year %d; %w", year, err)
		}

		if len(weather) > 0 {
			if err := s.db.InsertGameWeather(s.ctx, weather); err != nil {
				slog.Error("failed to insert game weather", "err", err)
				return fmt.Errorf("failed to insert game weather; %w", err)
			}
			totalInserted += len(weather)
			slog.Info(
				"inserted game weather",
				"year", int32ToString(year),
				"count", len(weather),
				"total", totalInserted,
			)
		}
	}

	slog.Info("game weather successfully inserted", "total_count", totalInserted)
	return nil
}

func (s *Seeder) SeedGameMedia() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		media, err := s.api.GetGameMedia(
			s.ctx, cfbd.GetGameMediaRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get game media",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf("failed to get game media for year %d; %w", year, err)
		}

		if len(media) > 0 {
			if err := s.db.InsertGameMedia(s.ctx, media); err != nil {
				slog.Error("failed to insert game media", "err", err)
				return fmt.Errorf("failed to insert game media; %w", err)
			}
			totalInserted += len(media)
			slog.Info(
				"inserted game media",
				"year", int32ToString(year),
				"count", len(media),
				"total", totalInserted,
			)
		}
	}

	slog.Info("game media successfully inserted", "total_count", totalInserted)
	return nil
}

func (s *Seeder) SeedBettingLines() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		lines, err := s.api.GetBettingLines(
			s.ctx, cfbd.GetBettingLinesRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get betting lines",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get betting lines for year %d; %w", year, err,
			)
		}

		if len(lines) > 0 {
			if err := s.db.InsertBettingLines(s.ctx, lines); err != nil {
				slog.Error("failed to insert betting lines", "err", err)
				return fmt.Errorf("failed to insert betting lines; %w", err)
			}
			totalInserted += len(lines)
			slog.Info(
				"inserted betting lines",
				"year", int32ToString(year),
				"count", len(lines),
				"total", totalInserted,
			)
		}
	}

	slog.Info("betting lines successfully inserted", "total_count", totalInserted)
	return nil
}

func (s *Seeder) SeedTeamRecords() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		records, err := s.api.GetTeamRecords(
			s.ctx, cfbd.GetTeamRecordsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get team records",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get team records for year %d; %w", year, err,
			)
		}

		if len(records) > 0 {
			if err := s.db.InsertTeamRecords(s.ctx, records); err != nil {
				slog.Error(
					"failed to insert team records",
					"year", int32ToString(year),
					"err", err,
				)

				return fmt.Errorf(
					"failed to insert team records; %w", err,
				)
			}

			totalInserted += len(records)
			slog.Info(
				"inserted team records",
				"year", int32ToString(year),
				"count", len(records),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"team records successfully inserted",
		"total_count", totalInserted,
	)
	return nil
}

func (s *Seeder) SeedTeamTalentComposite() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		talent, err := s.api.GetTeamTalentComposite(
			s.ctx, cfbd.GetTalentCompositeRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get team talent",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get team talent for year %d; %w", year, err,
			)
		}

		if len(talent) > 0 {
			if err := s.db.InsertTeamTalent(s.ctx, talent); err != nil {
				slog.Error(
					"failed to insert team talent",
					"year", int32ToString(year),
					"err", err,
				)

				return fmt.Errorf(
					"failed to insert team talent; %w", err,
				)
			}

			totalInserted += len(talent)
			slog.Info(
				"inserted team talent",
				"year", int32ToString(year),
				"count", len(talent),
				"total", totalInserted,
			)
		}
	}

	slog.Info("team talent successfully inserted", "total_count", totalInserted)
	return nil
}

func (s *Seeder) SeedTeamATS() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		ats, err := s.api.GetTeamATS(s.ctx, cfbd.GetTeamATSRequest{Year: year})
		if err != nil {
			slog.Error(
				"failed to get team ATS",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get team ATS for year %d; %w", year, err,
			)
		}

		if len(ats) > 0 {
			if err := s.db.InsertTeamATS(s.ctx, ats); err != nil {
				slog.Error("failed to insert team ATS", "err", err)
				return fmt.Errorf("failed to insert team ATS; %w", err)
			}

			totalInserted += len(ats)
			slog.Info(
				"inserted team ATS",
				"year", int32ToString(year),
				"count", len(ats),
				"total", totalInserted,
			)
		}
	}

	slog.Info("team ATS successfully inserted", "total_count", totalInserted)
	return nil
}

func (s *Seeder) SeedTeamSPPlus() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		ratings, err := s.api.GetTeamSPPlusRatings(
			s.ctx, cfbd.GetSPPlusRatingsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get team SP+ ratings",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get team SP+ ratings for year %d; %w", year, err,
			)
		}

		if len(ratings) > 0 {
			if err := s.db.InsertTeamSP(s.ctx, ratings); err != nil {
				slog.Error("failed to insert team SP+", "err", err)
				return fmt.Errorf("failed to insert team SP+; %w", err)
			}

			totalInserted += len(ratings)
			slog.Info(
				"inserted team SP+",
				"year", int32ToString(year),
				"count", len(ratings),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"team SP+ ratings successfully inserted",
		"total_count", totalInserted,
	)
	return nil
}

func (s *Seeder) SeedConferenceSPPlus() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		ratings, err := s.api.GetConferenceSPPlusRatings(
			s.ctx, cfbd.GetConferenceSPPlusRatingsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get conference SP+ ratings",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get conference SP+ ratings for year %d; %w", year, err,
			)
		}

		if len(ratings) > 0 {
			if err := s.db.InsertConferenceSP(s.ctx, ratings); err != nil {
				slog.Error("failed to insert conference SP+", "err", err)
				return fmt.Errorf("failed to insert conference SP+; %w", err)
			}

			totalInserted += len(ratings)
			slog.Info(
				"inserted conference SP+",
				"year", int32ToString(year),
				"count", len(ratings),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"conference SP+ ratings successfully inserted",
		"total_count", totalInserted,
	)

	return nil
}

func (s *Seeder) SeedTeamSRSRankings() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		ratings, err := s.api.GetSRSRatings(
			s.ctx, cfbd.GetSRSRatingsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get team SRS ratings",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get team SRS ratings for year %d; %w", year, err,
			)
		}

		if len(ratings) > 0 {
			if err := s.db.InsertTeamSRS(s.ctx, ratings); err != nil {
				slog.Error("failed to insert team SRS", "err", err)
				return fmt.Errorf("failed to insert team SRS; %w", err)
			}

			totalInserted += len(ratings)
			slog.Info(
				"inserted team SRS",
				"year", int32ToString(year),
				"count", len(ratings),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"team SRS ratings successfully inserted", "total_count", totalInserted,
	)
	return nil
}

func (s *Seeder) SeedTeamEloRankings() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		ratings, err := s.api.GetEloRatings(
			s.ctx, cfbd.GetEloRatingsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get team Elo ratings",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get team Elo ratings for year %d; %w", year, err,
			)
		}

		if len(ratings) > 0 {
			if err := s.db.InsertTeamElo(s.ctx, ratings); err != nil {
				slog.Error("failed to insert team Elo", "err", err)
				return fmt.Errorf("failed to insert team Elo; %w", err)
			}
			totalInserted += len(ratings)
			slog.Info(
				"inserted team Elo",
				"year", int32ToString(year),
				"count", len(ratings),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"team Elo ratings successfully inserted",
		"total_count", totalInserted,
	)

	return nil
}

func (s *Seeder) SeedTeamFPIRankings() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		ratings, err := s.api.GetFPIRatings(
			s.ctx, cfbd.GetFPIRatingsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get team FPI ratings",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get team FPI ratings for year %d; %w", year, err,
			)
		}

		if len(ratings) > 0 {
			if err := s.db.InsertTeamFPI(s.ctx, ratings); err != nil {
				slog.Error("failed to insert team FPI", "err", err)
				return fmt.Errorf("failed to insert team FPI; %w", err)
			}
			totalInserted += len(ratings)
			slog.Info(
				"inserted team FPI",
				"year", int32ToString(year),
				"count", len(ratings),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"team FPI ratings successfully inserted",
		"total_count", totalInserted,
	)
	return nil
}

func (s *Seeder) SeedWepaTeamSeason() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		metrics, err := s.api.GetTeamSeasonWEPA(
			s.ctx, cfbd.GetTeamSeasonWEPARequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get team season WEPA",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get team season WEPA for year %d; %w", year, err,
			)
		}

		if len(metrics) > 0 {
			if err := s.db.InsertAdjustedTeamMetrics(s.ctx, metrics); err != nil {
				slog.Error("failed to insert team season WEPA", "err", err)
				return fmt.Errorf("failed to insert team season WEPA; %w", err)
			}

			totalInserted += len(metrics)
			slog.Info(
				"inserted team season WEPA",
				"year", int32ToString(year),
				"count", len(metrics),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"team season WEPA successfully inserted",
		"total_count", totalInserted,
	)
	return nil
}

func (s *Seeder) SeedWepaPassing() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		wepa, err := s.api.GetPlayerPassingWEPA(
			s.ctx, cfbd.GetPlayerWEPARequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get passing WEPA",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get passing WEPA for year %d; %w", year, err,
			)
		}

		if len(wepa) > 0 {
			if err := s.db.InsertPlayerWeightedEPA(s.ctx, wepa); err != nil {
				slog.Error("failed to insert passing WEPA", "err", err)
				return fmt.Errorf("failed to insert passing WEPA; %w", err)
			}

			totalInserted += len(wepa)
			slog.Info(
				"inserted passing WEPA",
				"year", int32ToString(year),
				"count", len(wepa),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"passing WEPA successfully inserted",
		"total_count", totalInserted,
	)
	return nil
}

func (s *Seeder) SeedWepaRushing() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		wepa, err := s.api.GetPlayerRushingWEPA(
			s.ctx, cfbd.GetPlayerWEPARequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get rushing WEPA",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get rushing WEPA for year %d; %w", year, err,
			)
		}

		if len(wepa) > 0 {
			if err := s.db.InsertPlayerWeightedEPA(s.ctx, wepa); err != nil {
				slog.Error("failed to insert rushing WEPA", "err", err)
				return fmt.Errorf("failed to insert rushing WEPA; %w", err)
			}

			totalInserted += len(wepa)
			slog.Info(
				"inserted rushing WEPA",
				"year", int32ToString(year),
				"count", len(wepa),
				"total", totalInserted,
			)
		}
	}

	slog.Info("rushing WEPA successfully inserted", "total_count", totalInserted)
	return nil
}

func (s *Seeder) SeedWepaKicking() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		paar, err := s.api.GetPlayerKickingWEPA(
			s.ctx, cfbd.GetWepaPlayersKickingRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get kicking PAAR",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get kicking PAAR for year %d; %w", year, err,
			)
		}

		if len(paar) > 0 {
			if err := s.db.InsertKickerPAAR(s.ctx, paar); err != nil {
				slog.Error("failed to insert kicking PAAR", "err", err)
				return fmt.Errorf("failed to insert kicking PAAR; %w", err)
			}

			totalInserted += len(paar)
			slog.Info(
				"inserted kicking PAAR",
				"year", int32ToString(year),
				"count", len(paar),
				"total", totalInserted,
			)
		}
	}

	slog.Info("kicking PAAR successfully inserted", "total_count", totalInserted)
	return nil
}

func (s *Seeder) SeedReturningProduction() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		production, err := s.api.GetReturningProduction(
			s.ctx, cfbd.GetReturningProductionRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get returning production",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get returning production for year %d; %w", year, err,
			)
		}

		if len(production) > 0 {
			if err := s.db.InsertReturningProduction(s.ctx, production); err != nil {
				slog.Error("failed to insert returning production", "err", err)
				return fmt.Errorf("failed to insert returning production; %w", err)
			}

			totalInserted += len(production)
			slog.Info(
				"inserted returning production",
				"year", int32ToString(year),
				"count", len(production),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"returning production successfully inserted", "total_count", totalInserted,
	)
	return nil
}

func (s *Seeder) SeedPortalPlayers() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		players, err := s.api.GetTransferPortalPlayers(
			s.ctx, cfbd.GetTransferPortalPlayersRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get transfer portal players",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get transfer portal players for year %d; %w", year, err,
			)
		}

		if len(players) > 0 {
			if err := s.db.InsertPlayerTransfers(s.ctx, players); err != nil {
				slog.Error("failed to insert transfer portal players", "err", err)
				return fmt.Errorf("failed to insert transfer portal players; %w", err)
			}

			totalInserted += len(players)
			slog.Info(
				"inserted transfer portal players",
				"year", int32ToString(year),
				"count", len(players),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"transfer portal players successfully inserted",
		"total_count", totalInserted,
	)

	return nil
}

func (s *Seeder) SeedSeasonPlayerStats() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		stats, err := s.api.GetPlayerSeasonStats(
			s.ctx, cfbd.GetPlayerSeasonStatsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get player season stats",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get player season stats for year %d; %w", year, err,
			)
		}

		if len(stats) > 0 {
			if err := s.db.InsertPlayerStats(s.ctx, stats); err != nil {
				slog.Error("failed to insert player season stats", "err", err)
				return fmt.Errorf("failed to insert player season stats; %w", err)
			}

			totalInserted += len(stats)
			slog.Info(
				"inserted player season stats",
				"year", int32ToString(year),
				"count", len(stats),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"player season stats successfully inserted",
		"total_count", totalInserted,
	)

	return nil
}

func (s *Seeder) SeedSeasonTeamStats() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		stats, err := s.api.GetTeamSeasonStats(
			s.ctx, cfbd.GetTeamSeasonStatsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get team season stats",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get team season stats for year %d; %w", year, err,
			)
		}

		if len(stats) > 0 {
			if err := s.db.InsertTeamStats(s.ctx, stats); err != nil {
				slog.Error("failed to insert team season stats", "err", err)
				return fmt.Errorf("failed to insert team season stats; %w", err)
			}

			totalInserted += len(stats)
			slog.Info(
				"inserted team season stats",
				"year", int32ToString(year),
				"count", len(stats),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"team season stats successfully inserted",
		"total_count", totalInserted,
	)

	return nil
}

func (s *Seeder) SeedRankings() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		rankings, err := s.api.GetRankings(
			s.ctx, cfbd.GetRankingsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get rankings",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get rankings for year %d; %w", year, err,
			)
		}

		if len(rankings) > 0 {
			if err := s.db.InsertRankings(s.ctx, rankings); err != nil {
				slog.Error("failed to insert rankings", "err", err)
				return fmt.Errorf("failed to insert rankings; %w", err)
			}

			totalInserted += len(rankings)
			slog.Info(
				"inserted rankings",
				"year", int32ToString(year),
				"count", len(rankings),
				"total", totalInserted,
			)
		}
	}

	slog.Info("rankings successfully inserted", "total_count", totalInserted)
	return nil
}

func (s *Seeder) SeedRecruits() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		recruits, err := s.api.GetPlayerRecruitingRankings(
			s.ctx, cfbd.GetPlayersRecruitingRankingsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get recruits",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get recruits for year %d; %w", year, err,
			)
		}

		if len(recruits) > 0 {
			if err := s.db.InsertRecruits(s.ctx, recruits); err != nil {
				slog.Error("failed to insert recruits", "err", err)
				return fmt.Errorf("failed to insert recruits; %w", err)
			}

			totalInserted += len(recruits)
			slog.Info(
				"inserted recruits",
				"year", int32ToString(year),
				"count", len(recruits),
				"total", totalInserted,
			)
		}
	}

	slog.Info("recruits successfully inserted", "total_count", totalInserted)
	return nil
}

func (s *Seeder) SeedRecruitingRankings() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		rankings, err := s.api.GetTeamRecruitingRankings(
			s.ctx, cfbd.GetTeamRecruitingRankingsRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get recruiting rankings",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf(
				"failed to get recruiting rankings for year %d; %w", year, err,
			)
		}

		if len(rankings) > 0 {
			if err := s.db.InsertTeamRecruitingRankings(s.ctx, rankings); err != nil {
				slog.Error("failed to insert recruiting rankings", "err", err)
				return fmt.Errorf("failed to insert recruiting rankings; %w", err)
			}

			totalInserted += len(rankings)
			slog.Info(
				"inserted recruiting rankings",
				"year", int32ToString(year),
				"count", len(rankings),
				"total", totalInserted,
			)
		}
	}

	slog.Info(
		"recruiting rankings successfully inserted",
		"total_count", totalInserted,
	)
	return nil
}

func (s *Seeder) SeedDraftPicks() error {
	totalInserted := 0

	for _, year := range supportedYears {
		if err := s.throttle(s.ctx); err != nil {
			return fmt.Errorf("failed to wait for rate limit; %w", err)
		}

		picks, err := s.api.GetDraftPicks(
			s.ctx, cfbd.GetDraftPicksRequest{Year: year},
		)
		if err != nil {
			slog.Error(
				"failed to get draft picks",
				"year", int32ToString(year),
				"err", err,
			)

			return fmt.Errorf("failed to get draft picks for year %d; %w", year, err)
		}

		if len(picks) > 0 {
			if err := s.db.InsertDraftPicks(s.ctx, picks); err != nil {
				slog.Error("failed to insert draft picks", "err", err)
				return fmt.Errorf("failed to insert draft picks; %w", err)
			}

			totalInserted += len(picks)
			slog.Info(
				"inserted draft picks",
				"year", int32ToString(year),
				"count", len(picks),
				"total", totalInserted,
			)
		}
	}

	slog.Info("draft picks successfully inserted", "total_count", totalInserted)
	return nil
}

func int32ToString(val int32) string {
	return strconv.FormatInt(int64(val), 10)
}

package rss

import (
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/rs/zerolog/log"
	"time"
)

func StartWatcher(st *state.State) {
	log.Debug().Str("location", "feed watcher").Msg("starting feed watcher")
	go func() {
		timeUntilNextHour := time.Minute * time.Duration(60-time.Now().UTC().Minute())
		timeUntilNextHour += 30 * time.Second // little bit of buffer time to
		// make sure we're actually going to be within in the new hour

		log.Debug().Str("location", "feed watcher").Msgf("waiting %.2f minutes before starting ticker", timeUntilNextHour.Minutes())

		time.Sleep(timeUntilNextHour)

		runFeedProcessor(st, time.Now().UTC())

		log.Debug().Str("location", "feed watcher").Msg("starting ticker")

		ticker := time.NewTicker(time.Hour)
		for range ticker.C {
			// Yes, I am aware that you can get the current time from ticker.C
			// BUT that's been weird and caused some issues resulting in an
			// hour's task not being run, so I'm not using it
			runFeedProcessor(st, time.Now().UTC())
		}
	}()
}

func runFeedProcessor(st *state.State, currentTime time.Time) {
	currentTime = currentTime.UTC()
	log.Debug().
		Str("location", "feed watcher").
		Str("day", db.SendDayFromWeekday(currentTime.Weekday()).String()).
		Int("hour", currentTime.Hour()).
		Msg("running hourly job")
	if err := ProcessFeeds(st, db.SendDayFromWeekday(currentTime.Weekday()), currentTime.Hour()); err != nil {
		log.Error().Err(err).Str("location", "feed watcher").Send()
	}
}

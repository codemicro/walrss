package rss

import (
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/rs/zerolog/log"
	"time"
)

func StartWatcher(st *state.State) {
	go func() {
		currentTime := time.Now().UTC()
		
		timeUntilNextHour := time.Minute * time.Duration(60 - currentTime.Minute())
		timeUntilNextHour += 30 * time.Second // little bit of buffer time to
		// make sure we're actually going to be within in the new hour
		
		time.Sleep(timeUntilNextHour)

		runFeedProcessor(st, currentTime)

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
	log.Info().
		Str("location", "feed watcher").
		Str("day", db.SendDayFromWeekday(currentTime.Weekday()).String()).
		Int("hour", currentTime.Hour()).
		Msg("running hourly job")
	if err := ProcessFeeds(st, db.SendDayFromWeekday(currentTime.Weekday()), currentTime.Hour()); err != nil {
		log.Error().Err(err).Str("location", "feed watcher").Send()
	}
}

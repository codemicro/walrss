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
		time.Sleep(time.Minute * time.Duration(60-currentTime.Minute()))

		runFeedProcessor(st, currentTime)

		ticker := time.NewTicker(time.Hour)
		for currentTime := range ticker.C {
			runFeedProcessor(st, currentTime)
		}
	}()
}

func runFeedProcessor(st *state.State, currentTime time.Time) {
	log.Info().Str("location", "feed watcher").Msg("running hourly job")
	currentTime = currentTime.UTC()
	if err := ProcessFeeds(st, db.SendDayFromWeekday(currentTime.Weekday()), currentTime.Hour()); err != nil {
		log.Error().Err(err).Str("location", "feed watcher").Send()
	}
}

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

		if err := ProcessFeeds(st, db.SendDayFromWeekday(currentTime.Weekday()), currentTime.Hour()+1); err != nil {
			log.Error().Err(err).Str("location", "feed watcher").Send()
		}

		ticker := time.NewTicker(time.Hour)
		for currentTime := range ticker.C {
			if err := ProcessFeeds(st, db.SendDayFromWeekday(currentTime.Weekday()), currentTime.Hour()); err != nil {
				log.Error().Err(err).Str("location", "feed watcher").Send()
			}
		}
	}()
}

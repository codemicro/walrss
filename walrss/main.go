package main

import (
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/http"
	"github.com/codemicro/walrss/walrss/internal/rss"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/rs/zerolog/log"
	"os"
)

const dbFilename = "walrss.db"
const walrssDirectoryEnv = "WALRSS_DIR"

func run() error {
	if err := switchToDataDirectory(); err != nil {
		return err
	}

	st := state.New()
	if config, err := state.LoadConfig(); err != nil {
		return err
	} else {
		st.Config = config
	}

	store, err := db.New(dbFilename)
	if err != nil {
		return err
	}
	st.Data = store

	server, err := http.New(st)
	if err != nil {
		return err
	}

	rss.StartWatcher(st)

	return server.Run()
}

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("could not start")
	}
}

func switchToDataDirectory() error {
	if dir := os.Getenv(walrssDirectoryEnv); dir != "" {
		return os.Chdir(dir)
	}
	return nil
}

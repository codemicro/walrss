package db

import (
	"context"
	"embed"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"time"
)

var migs = migrate.NewMigrations()

//go:embed *.sql
var sqlMigrations embed.FS

func init() {
	if err := migs.Discover(sqlMigrations); err != nil {
		panic(err)
	}
}

func DoMigrations(db *bun.DB) error {
	log.Info().Msg("running migrations")

	mig := migrate.NewMigrator(db, migs)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := mig.Init(ctx); err != nil {
		return err
	}

	group, err := mig.Migrate(ctx)
	if err != nil {
		return err
	}

	if group.IsZero() {
		log.Info().Msg("database up to date")
	} else {
		log.Info().Msg("migrations applied")
	}

	return nil
}

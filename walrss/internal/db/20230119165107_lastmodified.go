package db

import (
	"context"
	"github.com/uptrace/bun"
	"strings"
)

func init() {
	migs.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			migLogger.Debug().Msg("2022-01-19@16:51:07 up")

			if _, err := db.ExecContext(ctx, `ALTER TABLE feeds ADD COLUMN last_modified VARCHAR;`); err != nil {
				if !strings.Contains(err.Error(), "duplicate column name") {
					return err
				}
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			migLogger.Debug().Msg("2022-01-19@16:51:07 down")

			if _, err := db.ExecContext(ctx, `ALTER TABLE feeds DROP COLUMN last_modified;`); err != nil {
				if !strings.Contains(err.Error(), "no such column") {
					return err
				}
			}

			return nil
		},
	)
}

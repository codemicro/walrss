package db

import (
	"context"
	"github.com/uptrace/bun"
	"strings"
)

func init() {
	migs.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			migLogger.Debug().Msg("2022-01-19@13:58:07 up")

			queries := []string{
				`ALTER TABLE feeds ADD COLUMN last_etag VARCHAR;`,
				`ALTER TABLE feeds ADD COLUMN cached_content VARCHAR;`,

			}

			for _, query := range queries {
				if _, err := db.ExecContext(ctx, query); err != nil {
					if !strings.Contains(err.Error(), "duplicate column name") {
						return err
					}
				}
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			migLogger.Debug().Msg("2022-01-19@13:58:07 down")

			queries := []string{
				`ALTER TABLE feeds DROP COLUMN last_etag;`,
				`ALTER TABLE feeds DROP COLUMN cached_content;`,

			}

			for _, query := range queries {
				if _, err := db.ExecContext(ctx, query); err != nil {
					if !strings.Contains(err.Error(), "no such column") {
						return err
					}
				}
			}

			return nil
		},
	)
}

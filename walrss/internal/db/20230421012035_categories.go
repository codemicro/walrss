package db

import (
	"context"
	"fmt"
	"github.com/uptrace/bun"
	"strings"
)

func init() {
	migs.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			migLogger.Debug().Msg("2022-04-21@01:20:35 up")

			if _, err := db.NewCreateTable().Model((*Category)(nil)).Exec(ctx); err != nil {
				return fmt.Errorf("creating category table: %w", err)
			}

			if _, err := db.ExecContext(ctx, `ALTER TABLE feeds ADD COLUMN category_id VARCHAR;`); err != nil {
				if !strings.Contains(err.Error(), "duplicate column name") {
					return fmt.Errorf("adding category_id to feeds: %w", err)
				}
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			migLogger.Debug().Msg("2022-04-21@01:20:35 down")

			if _, err := db.NewDropTable().Model((*Category)(nil)).Exec(ctx); err != nil {
				return fmt.Errorf("dropping category table: %w", err)
			}

			if _, err := db.ExecContext(ctx, `ALTER TABLE feeds DROP COLUMN category_id;`); err != nil {
				if !strings.Contains(err.Error(), "no such column") {
					return fmt.Errorf("removing category_id from feeds: %w", err)
				}
			}

			return nil
		},
	)
}

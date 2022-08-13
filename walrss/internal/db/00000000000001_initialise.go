package db

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	tps := []any{
		(*User)(nil),
		(*Feed)(nil),
	}

	migs.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			migLogger.Debug().Msg("1 up")

			for _, t := range tps {
				if _, err := db.NewCreateTable().Model(t).Exec(ctx); err != nil {
					return err
				}
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			migLogger.Debug().Msg("1 down")

			for _, t := range tps {
				if _, err := db.NewDropTable().Model(t).Exec(ctx); err != nil {
					return err
				}
			}

			return nil
		},
	)
}

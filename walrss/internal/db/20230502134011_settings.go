package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/uptrace/bun"
)

func init() {
	migs.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			migLogger.Debug().Msg("2023-05-02@13:40:11 up")

			if _, err := db.NewCreateTable().Model((*Settings)(nil)).Exec(ctx); err != nil {
				return fmt.Errorf("creating category table: %w", err)
			}

			rows, err := db.Query("SELECT name FROM pragma_table_info('users');")
			if err != nil {
				return err
			}

			var migrateUsers bool
			for rows.Next() {
				var x string
				if err := rows.Scan(&x); err != nil {
					return err
				}

				if x == "active" || x == "schedule_day" || x == "schedule_hour" {
					migrateUsers = true
					break
				}
			}
			_ = rows.Close()

			if migrateUsers {
				rows, err := db.Query("SELECT id, active, schedule_day, schedule_hour FROM users;")
				if err != nil {
					return err
				}

				var toInsert []*Settings

				for rows.Next() {
					var (
						userID       string
						active       bool
						scheduleDay  SendDay
						scheduleHour int
					)
					if err := rows.Scan(&userID, &active, &scheduleDay, &scheduleHour); err != nil {
						return err
					}

					toInsert = append(toInsert,
						&Settings{
							UserID:        userID,
							DigestsActive: active,
							ScheduleDay:   scheduleDay,
							ScheduleHour:  scheduleHour,
						},
					)
				}
				_ = rows.Close()

				if len(toInsert) != 0 {
					if _, err := db.NewInsert().Model(&toInsert).Exec(context.Background()); err != nil {
						return err
					}
				}
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			migLogger.Debug().Msg("2023-05-02@13:40:11 down")
			return errors.New("not implemented")
		},
	)
}

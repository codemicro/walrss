package core

import (
	"context"
	"database/sql"
	"errors"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
)

func GetUserSettings(st *state.State, userID string) (res *db.Settings, err error) {
	res = new(db.Settings)
	err = st.Data.NewSelect().Model(res).Where("id = ?", userID).Scan(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return
}

func UpdateSettings(st *state.State, settings *db.Settings) error {
	_, err := st.Data.NewUpdate().Model(settings).WherePK().Exec(context.Background())
	return err
}

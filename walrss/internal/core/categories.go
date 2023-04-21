package core

import (
	"context"
	"database/sql"
	"errors"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/lithammer/shortuuid/v4"
)

func NewCategory(st *state.State, userID, name string) (*db.Category, error) {
	if err := validateCategoryName(name); err != nil {
		return nil, err
	}

	category := &db.Category{
		ID:     shortuuid.New(),
		Name:   name,
		UserID: userID,
	}

	if _, err := st.Data.NewInsert().Model(category).Exec(context.Background()); err != nil {
		return nil, err
	}

	return category, nil
}

func GetCategory(st *state.State, categoryID string) (res *db.Category, err error) {
	res = new(db.Category)
	err = st.Data.NewSelect().
		Model(res).
		Relation("User").
		Where("Category.id = ?", categoryID).
		Scan(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return
}

func GetCategoriesForUser(st *state.State, userID string) (res []*db.Category, err error) {
	err = st.Data.NewSelect().
		Model(&res).
		Where("Category.user_id = ?", userID).
		Order("Category.name ASC").
		Scan(context.Background())
	return
}

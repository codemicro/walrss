package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func DeleteCategory(st *state.State, categoryID string) error {
	// move any feeds in this category to having no category
	_, err := st.Data.NewUpdate().
		Model((*db.Feed)(nil)).
		Set("category_id = NULL").
		Where("category_id = ?", categoryID).
		Exec(context.Background())
	if err != nil {
		return fmt.Errorf("DeleteCategory: deregistering feeds: %w, id=%s", err, categoryID)
	}

	_, err = st.Data.NewDelete().
		Model((*db.Category)(nil)).
		Where("id = ?", categoryID).
		Exec(context.Background())
	return err
}

package core

import (
	"context"
	"database/sql"
	"errors"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/lithammer/shortuuid/v4"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(st *state.State, email, password string) (*db.User, error) {
	if err := validateEmailAddress(email); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	u := &db.User{
		ID:    shortuuid.New(),
		Email: email,
		Salt:  generateRandomData(30),
	}

	hash, err := bcrypt.GenerateFromPassword(combineStringAndSalt(password, u.Salt), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u.Password = hash

	if _, err := st.Data.NewInsert().Model(u).Exec(context.Background()); err != nil {
		if e, ok := err.(*sqlite3.Error); ok {
			if e.Code == sqlite3.ErrConstraint {
				return nil, NewUserError("email address in use")
			}
		}
		return nil, err
	}

	return u, nil
}

func AreUserCredentialsCorrect(st *state.State, email, password string) (bool, error) {
	user, err := GetUserByEmail(st, email)
	if err != nil {
		return false, err
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, combineStringAndSalt(password, user.Salt)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func GetUserByID(st *state.State, userID string) (res *db.User, err error) {
	res = new(db.User)
	err = st.Data.NewSelect().Model(res).Where("id = ?", userID).Scan(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return
}

func GetUserByEmail(st *state.State, email string) (res *db.User, err error) {
	res = new(db.User)
	err = st.Data.NewSelect().Model(res).Where("email = ?", email).Scan(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return
}

func UpdateUser(st *state.State, user *db.User) error {
	_, err := st.Data.NewUpdate().Model(user).WherePK().Exec(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func GetUsersBySchedule(st *state.State, day db.SendDay, hour int) (res []*db.User, err error) {
	err = st.Data.NewSelect().
		Model(&res).
		Where(
			"active = ? and (schedule_day = ? or schedule_day = ?) and schedule_hour = ?",
			true,
			day,
			db.SendDaily,
			hour,
		).
		Scan(context.Background())
	return
}

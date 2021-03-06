package core

import (
	"errors"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/lithammer/shortuuid/v4"
	bh "github.com/timshannon/bolthold"
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

	if err := st.Data.Insert(u.ID, u); err != nil {
		if errors.Is(err, bh.ErrUniqueExists) {
			return nil, NewUserError("email address in use")
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

func GetUserByID(st *state.State, userID string) (*db.User, error) {
	user := new(db.User)
	if err := st.Data.FindOne(user, bh.Where("ID").Eq(userID)); err != nil {
		if errors.Is(err, bh.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func GetUserByEmail(st *state.State, email string) (*db.User, error) {
	user := new(db.User)
	if err := st.Data.FindOne(user, bh.Where("Email").Eq(email)); err != nil {
		if errors.Is(err, bh.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func UpdateUser(st *state.State, user *db.User) error {
	if err := st.Data.Update(user.ID, user); err != nil {
		if errors.Is(err, bh.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func GetUsersBySchedule(st *state.State, day db.SendDay, hour int) ([]*db.User, error) {
	// When trying to Or some queries, BH was weird, so it's easier to make two queries and combine them.
	// This ensures that indexes are used.

	var users []*db.User
	if err := st.Data.Find(&users,
		bh.Where("Schedule.Active").Eq(true).
			And("Schedule.Day").Eq(day).
			And("Schedule.Hour").Eq(hour),
	); err != nil {
		return nil, err
	}

	var users2 []*db.User
	if err := st.Data.Find(&users2,
		bh.Where("Schedule.Active").Eq(true).
			And("Schedule.Day").Eq(db.SendDaily).
			And("Schedule.Hour").Eq(hour),
	); err != nil {
		return nil, err
	}

	return append(users, users2...), nil
}

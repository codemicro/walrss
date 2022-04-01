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
	user := new(db.User)
	if err := st.Data.FindOne(user, bh.Where("Email").Eq(email)); err != nil {
		if errors.Is(err, bh.ErrNotFound) {
			return false, ErrNotFound
		}
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

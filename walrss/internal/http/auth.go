package http

import (
	"context"
	"errors"
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/views"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
	"github.com/stevelacy/daz"
	"math/rand"
	"sync"
	"time"
)

func (s *Server) authRegister(ctx *fiber.Ctx) error {

	if s.state.Config.Platform.DisableRegistration {
		ctx.Status(fiber.StatusForbidden)
		return views.SendPage(ctx, &views.PolyPage{
			TitleString: "Site registration disabled",
			BodyContent: daz.H("div",
				daz.Attr{"class": "container alert alert-danger"},
				"We're sorry - ",
				daz.H("b", "this instance of Walrss has registrations disabled"),
				". Please contact the operator of this Walrss instance with any queries.",
			)(),
		})
	}

	page := new(views.RegisterPage)

	if getCurrentUserID(ctx) != "" {
		goto success
	}

	if ctx.Method() == fiber.MethodPost {
		password := ctx.FormValue("password")
		passwordConfirmation := ctx.FormValue("passwordConfirmation")
		if password != passwordConfirmation {
			page.Problem = "Passwords do not match"
			goto exit
		}

		user, err := core.RegisterUser(
			s.state,
			ctx.FormValue("email"),
			password,
		)
		if err != nil {
			if core.IsUserError(err) {
				ctx.Status(core.GetUserErrorStatus(err))
				page.Problem = "Could not register account: " + err.Error()
				goto exit
			}
			return err
		}

		token := core.GenerateSessionToken(user.ID)

		ctx.Cookie(&fiber.Cookie{
			Name:     sessionCookieKey,
			Value:    token,
			Expires:  time.Now().UTC().Add(sessionDuration),
			Secure:   !s.state.Config.Debug,
			HTTPOnly: true,
		})

		goto success
	}

exit:
	return views.SendPage(ctx, page)
success:
	return ctx.Redirect(urls.Index)
}

func (s *Server) authSignIn(ctx *fiber.Ctx) error {
	page := &views.SignInPage{
		Problem:     ctx.Query("problem"),
		OIDCEnabled: s.state.Config.OIDC.Enable,
	}

	if getCurrentUserID(ctx) != "" {
		goto success
	}

	if ctx.Method() == fiber.MethodPost {
		email := ctx.FormValue("email")

		ok, err := core.AreUserCredentialsCorrect(
			s.state,
			email,
			ctx.FormValue("password"),
		)
		if err != nil {
			if errors.Is(err, core.ErrNotFound) {
				goto incorrectUsernameOrPassword
			}
			return err
		}

		if !ok {
			goto incorrectUsernameOrPassword
		}

		user, err := core.GetUserByEmail(s.state, email)
		if err != nil {
			return err
		}

		token := core.GenerateSessionToken(user.ID)

		ctx.Cookie(&fiber.Cookie{
			Name:     sessionCookieKey,
			Value:    token,
			Expires:  time.Now().UTC().Add(sessionDuration),
			Secure:   s.state.Config.EnableSecureCookies(),
			HTTPOnly: true,
		})

		goto success
	}

	return views.SendPage(ctx, page)
success:
	return ctx.Redirect(
		ctx.Query("next", urls.Index),
	)
incorrectUsernameOrPassword:
	ctx.Status(fiber.StatusUnauthorized)
	return views.SendPage(ctx, &views.SignInPage{Problem: "Incorrect username or password"})
}

var (
	knownStates = make(map[string]time.Time)
	stateLock   sync.Mutex
)

func init() {
	rand.Seed(time.Now().Unix())

	go func() {
		time.Sleep(time.Minute * 5)
		stateLock.Lock()

		var toDelete []string

		for k, v := range knownStates {
			if !v.After(time.Now().UTC()) {
				toDelete = append(toDelete, k)
			}
		}

		for _, k := range toDelete {
			delete(knownStates, k)
		}

		stateLock.Unlock()
	}()
}

func (s *Server) authOIDCOutbound(ctx *fiber.Ctx) error {
	if !s.state.Config.OIDC.Enable {
		return core.NewUserErrorWithStatus(fiber.StatusForbidden, "OIDC is disabled")
	}

	b := make([]byte, 30)
	for i := 0; i < len(b); i++ {
		b[i] = byte(65 + rand.Intn(25))
	}
	knownStates[string(b)] = time.Now().UTC().Add(time.Minute * 2)

	return ctx.Redirect(s.oauth2Config.AuthCodeURL(string(b)))
}

func (s *Server) authOIDCCallback(ctx *fiber.Ctx) error {
	if !s.state.Config.OIDC.Enable {
		return core.NewUserErrorWithStatus(fiber.StatusForbidden, "OIDC is disabled")
	}

	providedState := ctx.Query("state")
	stateLock.Lock()
	if exp, ok := knownStates[providedState]; ok && exp.After(time.Now().UTC()) {
		delete(knownStates, providedState)
		stateLock.Unlock()
	} else {
		stateLock.Unlock()
		return core.NewUserError("Invalid state")
	}

	oauth2Token, err := s.oauth2Config.Exchange(context.Background(), ctx.Query("code"))
	if err != nil {
		return err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return errors.New("missing ID token")
	}

	idToken, err := s.oidcVerifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		return err
	}

	var claims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return err
	}

	user, err := core.GetUserByEmail(s.state, claims.Email)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			if s.state.Config.Platform.DisableRegistration {
				return core.NewUserError("Cannot register user on-demand as registrations are disabled.")
			}
			user, err = core.RegisterUserOIDC(s.state, claims.Email)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	token := core.GenerateSessionToken(user.ID)

	ctx.Cookie(&fiber.Cookie{
		Name:     sessionCookieKey,
		Value:    token,
		Expires:  time.Now().UTC().Add(sessionDuration),
		Secure:   s.state.Config.EnableSecureCookies(),
		HTTPOnly: true,
	})

	return ctx.Redirect(urls.Index)
}

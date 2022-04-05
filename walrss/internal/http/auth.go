package http

import (
	"errors"
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/views"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
	"github.com/stevelacy/daz"
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
		Problem: ctx.Query("problem"),
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

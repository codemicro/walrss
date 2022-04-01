package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/gofiber/fiber/v2"
	"time"
)

func (s *Server) authRegister(ctx *fiber.Ctx) error {
	user, err := core.RegisterUser(s.state,
		ctx.FormValue("email"),
		ctx.FormValue("password"),
	)
	if err != nil {
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

	return ctx.SendString("ok!")
}

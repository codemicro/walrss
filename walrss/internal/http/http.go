package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	sessionCookieKey = "walrss-session"
	sessionDuration  = (time.Hour * 24) * 7 // 7 days
)

type Server struct {
	state *state.State
	app   *fiber.App
}

func New(st *state.State) (*Server, error) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: !st.Config.Debug,
		AppName:               "Walrss",
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "Internal Server Error"

			switch e := err.(type) {
			case *fiber.Error:
				code = e.Code
				msg = err.Error()
			case *core.UserError:
				code = e.Status
				msg = err.Error()
			default:
				log.Error().Err(err).Str("location", "http").Str("url", ctx.OriginalURL()).Send()
			}

			ctx.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
			return ctx.Status(code).SendString(msg)
		},
	})

	s := &Server{
		state: st,
		app:   app,
	}

	s.registerHandlers()

	return s, nil
}

func (s *Server) registerHandlers() {
	s.app.Get(urls.AuthRegister, s.authRegister)
	s.app.Post(urls.AuthRegister, s.authRegister)

	s.app.Get(urls.AuthSignIn, s.authSignIn)
	s.app.Post(urls.AuthSignIn, s.authSignIn)
}

func (s *Server) Run() error {
	return s.app.Listen(s.state.Config.GetHTTPAddress())
}

func UserErrorToResponse(ctx *fiber.Ctx, ue core.UserError) error {
	ctx.Status(ue.Status)
	return ctx.SendString(ue.Error())
}

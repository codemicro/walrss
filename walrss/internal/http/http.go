package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
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
	})
	// TODO: Add error handler with UserError support

	s := &Server{
		state: st,
		app:   app,
	}

	s.registerHandlers()

	return s, nil
}

func (s *Server) registerHandlers() {
	s.app.Post(urls.AuthRegister, s.authRegister)
}

func (s *Server) Run() error {
	return s.app.Listen(s.state.Config.GetHTTPAddress())
}

func UserErrorToResponse(ctx *fiber.Ctx, ue core.UserError) error {
	ctx.Status(ue.Status)
	return ctx.SendString(ue.Error())
}

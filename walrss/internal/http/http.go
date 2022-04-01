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
	userIDLocalKey   = "userID"
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
	s.app.Use(func(ctx *fiber.Ctx) error {
		if token := ctx.Cookies(sessionCookieKey); token != "" {
			log.Debug().Msgf("cookie %s=%s", sessionCookieKey, token)
			userID, createdAt, err := core.ValidateSessionToken(token)
			if err == nil && time.Now().Sub(createdAt) < sessionDuration {
				log.Debug().Msg("session valid")
				ctx.Locals(userIDLocalKey, userID)
			}
		}

		return ctx.Next()
	})

	s.app.Get(urls.AuthRegister, s.authRegister)
	s.app.Post(urls.AuthRegister, s.authRegister)

	s.app.Get(urls.AuthSignIn, s.authSignIn)
	s.app.Post(urls.AuthSignIn, s.authSignIn)
}

func (s *Server) Run() error {
	return s.app.Listen(s.state.Config.GetHTTPAddress())
}

func userErrorToResponse(ctx *fiber.Ctx, ue core.UserError) error {
	ctx.Status(ue.Status)
	return ctx.SendString(ue.Error())
}

func getCurrentUserID(ctx *fiber.Ctx) string {
	if x := ctx.Locals(userIDLocalKey); x != nil {
		s, ok := x.(string)
		if ok {
			return s
		}
	}
	return ""
}

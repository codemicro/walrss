package http

import (
	"context"
	"errors"
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/views"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/codemicro/walrss/walrss/internal/static"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/coreos/go-oidc"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/stevelacy/daz"
	"golang.org/x/oauth2"
	"net/url"
	"strings"
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

	oidcProvider *oidc.Provider
	oidcVerifier *oidc.IDTokenVerifier
	oauth2Config *oauth2.Config
}

func New(st *state.State) (*Server, error) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: !st.Config.Debug,
		AppName:               "Walrss",
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "Internal Server Error"

			var (
				fiberErr *fiber.Error
				userErr  *core.UserError
			)

			if errors.As(err, &fiberErr) {
				code = fiberErr.Code
				msg = fiberErr.Error()
			} else if errors.As(err, &userErr) {
				code = userErr.Status
				msg = userErr.Error()
			} else {
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

	if st.Config.OIDC.Enable {
		provider, err := oidc.NewProvider(context.Background(), st.Config.OIDC.Issuer)
		if err != nil {
			return nil, err
		}

		s.oidcProvider = provider
		s.oidcVerifier = provider.Verifier(&oidc.Config{ClientID: st.Config.OIDC.ClientID})
		s.oauth2Config = &oauth2.Config{
			ClientID:     st.Config.OIDC.ClientID,
			ClientSecret: st.Config.OIDC.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  strings.TrimSuffix(st.Config.Server.ExternalURL, "/") + urls.AuthOIDCCallback,
			Scopes:       []string{"email", "profile", "openid"},
		}
	}

	s.registerHandlers()

	return s, nil
}

func (s *Server) registerHandlers() {
	s.app.Use(func(ctx *fiber.Ctx) error {
		if token := ctx.Cookies(sessionCookieKey); token != "" {
			userID, createdAt, err := core.ValidateSessionToken(token)
			if err == nil && time.Now().Sub(createdAt) < sessionDuration {
				ctx.Locals(userIDLocalKey, userID)
			}
		}

		return ctx.Next()
	})

	s.app.Get(urls.Index, s.mainPage)

	s.app.Get(urls.AuthRegister, s.authRegister)
	s.app.Post(urls.AuthRegister, s.authRegister)

	s.app.Get(urls.AuthSignIn, s.authSignIn)
	s.app.Post(urls.AuthSignIn, s.authSignIn)

	s.app.Get(urls.AuthOIDCOutbound, s.authOIDCOutbound)
	s.app.Get(urls.AuthOIDCCallback, s.authOIDCCallback)

	s.app.Put(urls.EditEnabledState, s.editEnabledState)
	s.app.Put(urls.EditTimings, s.editTimings)

	s.app.Get(urls.EditFeedItem, s.editFeedItem)
	s.app.Put(urls.EditFeedItem, s.editFeedItem)
	s.app.Delete(urls.EditFeedItem, s.editFeedItem)
	s.app.Get(urls.CancelEditFeedItem, s.cancelEditFeedItem)

	s.app.Get(urls.NewFeedItem, s.newFeedItem)
	s.app.Post(urls.NewFeedItem, s.newFeedItem)

	s.app.Post(urls.SendTestEmail, s.sendTestEmail)
	s.app.Get(urls.TestEmailStatus, s.testEmailStatus)

	s.app.Get(urls.ExportAsOPML, s.exportAsOPML)
	s.app.Post(urls.ImportFromOPML, s.importFromOPML)

	s.app.Use(urls.Statics, static.NewHandler())
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

func requestStandardSignIn(ctx *fiber.Ctx) error {
	rdu := ctx.OriginalURL() // TODO: Could this use of OriginalURL be insecure?

	queryParams := make(url.Values)
	queryParams.Add("problem", "Please sign in first.")
	queryParams.Add("next", rdu)
	nextURL := urls.AuthSignIn + "?" + queryParams.Encode()

	ctx.Status(fiber.StatusUnauthorized)

	// Instead of plainly redirecting, we use a HTML redirect here. This is to clear the HTTP verb used for this
	// request. For example - if the request was made with DELETE, using ctx.Redirect will preserve that verb. Using
	// this method will restart with a GET verb.
	return views.SendPage(ctx, &views.PolyPage{
		TitleString:      "Please sign in first",
		BodyContent:      daz.H("p", "Please sign in first. If your browser doesn't automatically redirect you, click ", daz.H("a", daz.Attr{"href": nextURL}, "here"), ".")(),
		ExtraHeadContent: daz.H("meta", daz.Attr{"http-equiv": "Refresh", "content": "0; " + nextURL})(),
	})
}

func requestFragmentSignIn(ctx *fiber.Ctx, nextURL string) error {
	queryParams := make(url.Values)
	queryParams.Add("problem", "Please sign in first.")
	queryParams.Add("next", nextURL)

	ctx.Set("HX-Redirect", urls.AuthSignIn+"?"+queryParams.Encode())
	return nil
}

func fragmentEmitSuccess(ctx *fiber.Ctx) {
	ctx.Set("HX-Trigger", "successResponse")
}

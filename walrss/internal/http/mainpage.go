package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/neoviews"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) mainPage(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestStandardSignIn(ctx)
	}

	user, err := core.GetUserByID(s.state, currentUserID)
	if err != nil {
		return err
	}

	feeds, err := core.GetFeedsForUser(s.state, currentUserID)
	if err != nil {
		return err
	}

	ctx.Type("html")
	return ctx.SendString(neoviews.FeedsPage(&neoviews.FeedsPageArgs{
		DigestsEnabled: user.Active,
		Feeds:          feeds,
	}))
}

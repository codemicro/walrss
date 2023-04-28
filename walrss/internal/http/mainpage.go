package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/neoviews"
	"github.com/codemicro/walrss/walrss/internal/urls"
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

	var x string
	feeds, err := core.GetFeeds(s.state, &core.GetFeedsArgs{UserID: currentUserID, CategoryID: &x})
	if err != nil {
		return err
	}

	cats, err := core.GetCategoriesForUser(s.state, currentUserID)
	if err != nil {
		return err
	}

	ctx.Type("html")
	return ctx.SendString(neoviews.FeedsPage(&neoviews.FeedsPageArgs{
		DigestsEnabled: user.Active,
		Feeds:          feeds,
		Categories:     cats,
	}))
}

func (s *Server) getFeedsTab(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}
	category := ctx.Query("category")

	resp, err := neoviews.RenderFeedTabsAndTableForUser(s.state, currentUserID, category, false)
	if err != nil {
		return err
	}
	return ctx.SendString(resp)
}
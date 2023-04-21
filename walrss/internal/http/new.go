package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/neoviews"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
	"github.com/stevelacy/daz"
)

func (s *Server) newFeedItem(ctx *fiber.Ctx) error {

	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	switch ctx.Method() {
	case fiber.MethodGet:
		return ctx.SendString(neoviews.FragmentNewFeed())
		//return ctx.SendString(views.RenderNewFeedItemRow())
	case fiber.MethodPost:
		_, err := core.NewFeed(
			s.state,
			currentUserID,
			ctx.FormValue("name"),
			ctx.FormValue("url"),
		)
		if err != nil {
			return err
		}

		// TODO: Set category
		resp, err := neoviews.RenderFeedTabsAndTableForUser(s.state, currentUserID, "", true)
		if err != nil {
			return err
		}
		fragmentEmitSuccess(ctx)
		return ctx.SendString(daz.H("div")() + resp)
	}

	panic("unreachable")
}

func (s *Server) newCategory(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	switch ctx.Method() {
	case fiber.MethodGet:
		return ctx.SendString(neoviews.FragmentNewCategory())
	case fiber.MethodPost:
		newCat, err := core.NewCategory(
			s.state,
			currentUserID,
			ctx.FormValue("name"),
		)
		if err != nil {
			return err
		}

		resp, err := neoviews.RenderFeedTabsAndTableForUser(s.state, currentUserID, newCat.ID, true)
		if err != nil {
			return err
		}
		fragmentEmitSuccess(ctx)
		return ctx.SendString(daz.H("div")() + resp)
	}

	panic("unreachable")
}

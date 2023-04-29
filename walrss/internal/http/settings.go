package http

import (
	"github.com/codemicro/walrss/walrss/internal/http/neoviews"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) settingsPage(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestStandardSignIn(ctx)
	}
	ctx.Type("html")
	return ctx.SendString(neoviews.SettingsPage())
}

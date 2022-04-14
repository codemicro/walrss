package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) exportAsOPML(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestStandardSignIn(ctx)
	}

	exported, err := core.ExportFeedsForUser(s.state, currentUserID)
	if err != nil {
		return err
	}

	ctx.Set(fiber.HeaderContentType, "application/xml")
	return ctx.Send(exported)
}

package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/rss"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func (s *Server) sendTestEmail(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	user, err := core.GetUserByID(s.state, currentUserID)
	if err != nil {
		return err
	}

	go func() {
		if err := rss.ProcessUserFeed(s.state, user); err != nil {
			log.Error().Err(err).Str("location", "test email").Str("user", user.ID).Send()
		}
	}()

	fragmentEmitSuccess(ctx)
	return nil
}

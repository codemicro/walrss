package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/views"
	"github.com/codemicro/walrss/walrss/internal/rss"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"sync"
	"time"
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

	testEmailStatesLock.Lock()
	testEmailStates[currentUserID] = testEmailStates[currentUserID] + "Starting at " + time.Now().UTC().Format(time.RFC1123Z) + "\n"
	testEmailStatesLock.Unlock()

	go func() {
		status := make(chan string, 50)

		var err error
		go func() {
			err = rss.ProcessUserFeed(s.state, user, status)
		}()

		for statusAddition := range status {
			testEmailStatesLock.Lock()
			testEmailStates[currentUserID] = testEmailStates[currentUserID] + statusAddition + "\n"
			testEmailStatesLock.Unlock()
		}

		if err != nil {
			log.Error().Err(err).Str("location", "test email").Str("user", user.ID).Send()
		}
	}()

	return s.testEmailStatus(ctx)
}

var (
	testEmailStates     = make(map[string]string)
	testEmailStatesLock sync.RWMutex
)

func (s *Server) testEmailStatus(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	end, _ := strconv.ParseBool(ctx.Query("end", "false"))

	testEmailStatesLock.RLock()
	var content string
	if end {
		delete(testEmailStates, currentUserID)
	} else {
		content = testEmailStates[currentUserID]
	}
	defer testEmailStatesLock.RUnlock()

	if end {
		ctx.Set("HX-Refresh", "true")
		return nil
	}

	var endOnNext bool
	if strings.HasSuffix(strings.TrimSpace(content), "Done!") {
		endOnNext = true
		fragmentEmitSuccess(ctx)
	}
	return ctx.SendString(views.RenderTestEmailBox(content, endOnNext))
}

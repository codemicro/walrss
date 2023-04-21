package neoviews

import (
	"embed"
	"github.com/codemicro/walrss/walrss/internal/http/neoviews/internal/components"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"net/http"
)

//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -skipLineComments -ext qtpl.html

func SendPage(ctx *fiber.Ctx) error {
	ctx.Set(fiber.HeaderContentType, "html")
	return nil
}

var (
	menuItemFeeds = &components.MenuItem{
		Path: "/feeds",
		Text: "My Feeds",
		Icon: "bi-rss-fill",
	}
	menuItemSettings = &components.MenuItem{
		Path: "/settings",
		Text: "Settings",
		Icon: "bi-gear-fill",
	}
	menuItemAccount = &components.MenuItem{
		Path: "/account",
		Text: "Account (INOP)",
		Icon: "bi-person-fill",
	}
	menuItemSignOut = &components.MenuItem{
		Path: "/auth/signout",
		Text: "Sign Out (INOP)",
		Icon: "bi-door-open-fill",
	}

	menuItems = []*components.MenuItem{
		menuItemFeeds,
		menuItemSettings,
		menuItemAccount,
		menuItemSignOut,
	}
)

//go:embed static/*
var statics embed.FS

func GetStaticHandler() fiber.Handler {
	return filesystem.New(filesystem.Config{
		Root:       http.FS(statics),
		PathPrefix: "static",
	})
}

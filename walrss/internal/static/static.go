package static

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed assets
var assets embed.FS

func NewHandler() fiber.Handler {
	return filesystem.New(filesystem.Config{
		Root:       http.FS(assets),
		PathPrefix: "assets",
	})
}

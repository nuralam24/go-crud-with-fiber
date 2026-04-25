package http

import (
	_ "embed"

	"github.com/gofiber/fiber/v2"
)

//go:embed swagger_openapi.yaml
var openAPIYAML []byte

func RegisterSwaggerRoutes(app *fiber.App) {
	app.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/yaml; charset=utf-8")
		return c.Send(openAPIYAML)
	})

	app.Get("/swagger", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>Go CRUD API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
      });
    };
  </script>
</body>
</html>`)
	})
}

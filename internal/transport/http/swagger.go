package http

import (
	_ "embed"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

//go:embed swagger_openapi.yaml
var openAPIYAML []byte

// Public doc URLs (shared convention with Gin service).
const (
	SwaggerUIPath   = "/api/docs"
	OpenAPISpecPath = "/api/docs/openapi.yaml"
)

func RegisterSwaggerRoutes(app *fiber.App) {
	serveSpec := func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/yaml; charset=utf-8")
		return c.Send(openAPIYAML)
	}
	serveUI := func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(swaggerUIHTML(OpenAPISpecPath))
	}

	app.Get(OpenAPISpecPath, serveSpec)
	app.Get(OpenAPISpecPath+"/", func(c *fiber.Ctx) error {
		return c.Redirect(OpenAPISpecPath)
	})

	app.Get(SwaggerUIPath, serveUI)
	app.Get(SwaggerUIPath+"/", func(c *fiber.Ctx) error {
		return c.Redirect(SwaggerUIPath)
	})

	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.Redirect(SwaggerUIPath)
	})
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect(SwaggerUIPath)
	})
	app.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		return c.Redirect(OpenAPISpecPath)
	})
}

func swaggerUIHTML(specURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
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
        url: %q,
        dom_id: "#swagger-ui",
      });
    };
  </script>
</body>
</html>`, specURL)
}

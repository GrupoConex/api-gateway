package main

import (
	"log"

	"github.com/fibex/gateway/pkg/auth"
	"github.com/fibex/gateway/pkg/config"
	"github.com/fibex/gateway/pkg/proxy"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	fiberProxy "github.com/gofiber/fiber/v2/middleware/proxy"
)

func main() {
	cfg := config.LoadConfig()

	app := fiber.New(fiber.Config{
		AppName: "Fibex API Gateway v2.0 (SOLID)",
	})

	app.Use(logger.New())

	authenticator := auth.NewKeycloakAuthenticator(cfg.KeycloakURL, cfg.KeycloakRealm)
	router := proxy.NewFiberRouter(cfg.Routes)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "UP", "version": "2.0-solid"})
	})

	api := app.Group("/api")
	
	
	api.Post("/v1/employees", func(c *fiber.Ctx) error {
		intranetURL := cfg.Routes["intranet"]
		if intranetURL == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Intranet route not configured"})
		}
		return fiberProxy.Do(c, intranetURL+"/api/employees/profix-data")
	})

	api.Use(authenticator.Middleware())
	api.All("/*", router.Handle)

	log.Fatal(app.Listen(":" + cfg.Port))
}

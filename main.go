package main

import (
	"log"
	"strings"

	"github.com/fibex/gateway/pkg/auth"
	"github.com/fibex/gateway/pkg/config"
	"github.com/fibex/gateway/pkg/proxy"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	fiberProxy "github.com/gofiber/fiber/v2/middleware/proxy"
)

func main() {
	cfg := config.LoadConfig()

	app := fiber.New(fiber.Config{
		AppName: "Fibex API Gateway v2.0 (SOLID)",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // En producción podemos restringirlo más luego
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Cookie",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		AllowCredentials: true,
	}))

	authenticator := auth.NewKeycloakAuthenticator(cfg.KeycloakURL, cfg.KeycloakRealm)
	router := proxy.NewFiberRouter(cfg.Routes)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "UP", "version": "2.0-solid"})
	})

	api := app.Group("/api")

	// --- RUTAS PÚBLICAS DE AUTENTICACIÓN (VIÁTICOS) ---
	// Estas rutas se definen ANTES del middleware para que Keycloak y el Frontend puedan acceder sin token.
	
	// Discovery OIDC
	api.Get("/v1/viaticos/auth/.well-known/openid-configuration", func(c *fiber.Ctx) error {
		viaticosURL := strings.TrimSuffix(cfg.Routes["viaticos"], "/")
		if viaticosURL == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Viaticos route not configured"})
		}
		targetURL := viaticosURL + "/auth/.well-known/openid-configuration"
		return fiberProxy.Do(c, targetURL)
	})

	// JWKS para validación de firmas
	api.Get("/v1/viaticos/auth/jwks", func(c *fiber.Ctx) error {
		viaticosURL := strings.TrimSuffix(cfg.Routes["viaticos"], "/")
		if viaticosURL == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Viaticos route not configured"})
		}
		targetURL := viaticosURL + "/auth/jwks"
		return fiberProxy.Do(c, targetURL)
	})

	// Proxy para Login/Session/OIDC Flow
	api.All("/v1/viaticos/auth/*", func(c *fiber.Ctx) error {
		viaticosURL := strings.TrimSuffix(cfg.Routes["viaticos"], "/")
		if viaticosURL == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Viaticos route not configured"})
		}
		path := c.Params("*")
		targetURL := viaticosURL + "/auth/" + path
		return fiberProxy.Do(c, targetURL)
	})

	api.Use(authenticator.Middleware())
	api.All("/*", router.Handle)

	log.Fatal(app.Listen(":" + cfg.Port))
}

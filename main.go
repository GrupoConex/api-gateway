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
		AllowOriginsFunc: func(origin string) bool {
			// Permitimos cualquier origen en desarrollo y producción para VIATIX
			return true 
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Cookie, X-Requested-With, X-Framework",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		AllowCredentials: true,
		ExposeHeaders:    "Set-Cookie",
	}))

	authenticator := auth.NewKeycloakAuthenticator(cfg.KeycloakURL, cfg.KeycloakRealm)
	router := proxy.NewFiberRouter(cfg.Routes)
	
	log.Println("[CONFIG] Cargando rutas de proxy...")
	for svc, url := range cfg.Routes {
		log.Printf("[CONFIG] %s -> %s", svc, url)
	}

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
		if !strings.HasPrefix(viaticosURL, "http") {
			viaticosURL = "https://" + viaticosURL
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
		if !strings.HasPrefix(viaticosURL, "http") {
			viaticosURL = "https://" + viaticosURL
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
		
		// Blindaje: Asegurar protocolo
		if !strings.HasPrefix(viaticosURL, "http") {
			viaticosURL = "https://" + viaticosURL
		}

		path := c.Params("*")
		targetURL := viaticosURL + "/auth/" + path
		
		log.Printf("[PROXY] Viaticos Auth: %s %s -> %s", c.Method(), c.Path(), targetURL)
		
		// Inyectamos cabeceras para que Better-Auth sepa que viene del Gateway
		c.Request().Header.Set("X-Forwarded-Host", c.Hostname())
		c.Request().Header.Set("X-Forwarded-Proto", c.Protocol())
		
		return fiberProxy.Do(c, targetURL)
	})

	api.Use(authenticator.Middleware())
	api.All("/*", router.Handle)

	log.Fatal(app.Listen(":" + cfg.Port))
}

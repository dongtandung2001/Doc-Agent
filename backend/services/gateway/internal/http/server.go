package http

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1/protov1connect"
	"github.com/dongtandung2001/Doc-Agent/backend/services/gateway/internal/config"
	"github.com/dongtandung2001/Doc-Agent/backend/services/gateway/internal/handlers"
)

type Server struct {
	app    *fiber.App
	config *config.Config
}

func NewServer(
	cfg *config.Config,
	gatewayHandler *handlers.GatewayHandler,
) *Server {
	app := fiber.New(fiber.Config{
		AppName: "Doc-Agent API Gateway",
	})

	// Apply shared middleware
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "gateway",
		})
	})

	// Mount Connect-Go GatewayService handler
	// All Connect handlers are in the protov1connect package
	gatewayPath, gatewayHdlr := protov1connect.NewGatewayServiceHandler(gatewayHandler)

	// Mount the gateway handler to respond to all Connect-Go requests
	app.All(gatewayPath+"*", adaptor.HTTPHandler(gatewayHdlr))

	return &Server{
		app:    app,
		config: cfg,
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	log.Printf("🚀 Gateway starting on %s", addr)
	return s.app.Listen(addr)
}

func (s *Server) Shutdown() error {
	log.Println("Gateway shutting down...")
	return s.app.Shutdown()
}

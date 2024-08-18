package server

import (
	"bootcamp_task/cache"
	"bootcamp_task/config"
	"bootcamp_task/handlers"
	"bootcamp_task/storage/storages"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"go.uber.org/fx"
	"strconv"
)

func buildFiberServer(lc fx.Lifecycle, h *handlers.Handlers, c *config.Config) *fiber.App {
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Get("/dummyLogin", h.DummyLogin)
	app.Post("/register", h.Register)
	app.Post("/login", h.Login)

	houseGroup := app.Group("/house")
	houseGroup.Post("/create", h.CreateHome)
	houseGroup.Get("/:id", h.GetHouseFlats)

	flatsGroup := app.Group("/flat")
	flatsGroup.Post("/create", h.CreateFlat)
	flatsGroup.Post("/update", h.UpdateFlat)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go app.Listen(":" + strconv.Itoa(c.ServerPort))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown()
		},
	})

	return app
}

func BuildServerAndEnv() *fx.App {
	return fx.New(
		fx.Provide(
			config.ParseConfig,
			cache.NewCache,
			storages.NewStorage,
			handlers.NewHandlers,
		),
		fx.Invoke(buildFiberServer),
	)
}

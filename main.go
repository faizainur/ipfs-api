package main

import (
	ipfs "github.com/faizainur/ipfs-api/ipfs_client"
	"github.com/faizainur/ipfs-api/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(middleware)

	IpfsMiddleware := middlewares.IpfsMiddleware{
		&ipfs.IPFSClient{
			ApiServerUri:     "http://127.0.0.1:5001/api/v0/",
			GatewayServerUri: "http://localhost:7000/ipfs/",
		},
	}

	app.Get("/ping", ping)
	app.Post("/upload", IpfsMiddleware.UploadFile)
	app.Post("/upload2", Upload)

	app.Listen(":4000")
}

func ping(c *fiber.Ctx) error {
	return c.JSON(map[string]interface{}{
		"code":    200,
		"status":  "server is running",
		"message": c.Locals("vals"),
	})
}

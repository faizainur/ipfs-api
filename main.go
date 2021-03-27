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

	ipfsMiddleware := middlewares.IpfsMiddleware{
		IpfsClient: &ipfs.IPFSClient{
			ApiServerUri:     "http://127.0.0.1:5001/api/v0/",
			GatewayServerUri: "http://127.0.0.1:7000/ipfs/",
		},
	}

	v1 := app.Group("/v1")
	{
		v1.Get("/ping", ping)

		ipfs := v1.Group("/ipfs")
		{
			ipfs.Get("/fetch", ipfsMiddleware.FetchFile)
			ipfs.Post("/upload", ipfsMiddleware.FetchFile)
		}
	}

	app.Listen(":4000")
}

func ping(c *fiber.Ctx) error {
	return c.JSON(map[string]interface{}{
		"code":    200,
		"status":  "server is running",
		"message": c.Locals("vals"),
	})
}

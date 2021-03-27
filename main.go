package main

import (
	"io/ioutil"

	ipfs "github.com/faizainur/ipfs-api/ipfs_client"
	"github.com/faizainur/ipfs-api/middlewares"
	"github.com/faizainur/ipfs-api/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(middleware)

	cryptoService := services.NewCryptoService(loadKey())
	ipfsClient := ipfs.NewClient("http://127.0.0.1:5001/api/v0/", "http://127.0.0.1:7000/ipfs/")

	ipfsMiddleware := middlewares.IpfsMiddleware{
		IpfsClient:    ipfsClient,
		CryptoService: cryptoService,
	}

	v1 := app.Group("/v1")
	{
		v1.Get("/ping", ping)

		ipfs := v1.Group("/ipfs")
		{
			ipfs.Get("/fetch", ipfsMiddleware.FetchFile)
			ipfs.Post("/upload", ipfsMiddleware.UploadFile)
		}
	}

	app.Listen(":4000")
}

func loadKey() []byte {
	key, err := ioutil.ReadFile("secret.key")
	if err != nil {
		panic(err)
	}
	return key
}

func ping(c *fiber.Ctx) error {
	return c.JSON(map[string]interface{}{
		"code":    200,
		"status":  "server is running",
		"message": c.Locals("vals"),
	})
}

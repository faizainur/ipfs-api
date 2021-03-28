package main

import (
	"encoding/hex"
	"io/ioutil"

	"github.com/faizainur/ipfs-api/cutils"
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

	// cutils.GenerateKeyFile()

	cryptoService := services.NewCryptoService(loadKey())
	authService := services.NewAuthService("http://localhost:8000/v1/auth/validate_token", "localhost:9001")
	ipfsClient := ipfs.NewClient("http://127.0.0.1:5001/api/v0/", "http://127.0.0.1:7000/ipfs/")

	ipfsMiddleware := middlewares.IpfsMiddleware{
		IpfsClient:    ipfsClient,
		CryptoService: cryptoService,
	}

	authMiddleware := middlewares.AuthMiddleware{
		AuthService: authService,
	}

	v1 := app.Group("/v1")
	{
		v1.Get("/ping", ping)
		v1.Get("/secure", authMiddleware.ValidateJwtToken, securedEndpoint)
		v1.Get("/secureOauth", authMiddleware.IntrospectOauth2Token, securedEndpoint)

		ipfs := v1.Group("/ipfs")
		{
			ipfs.Get("/fetch", authMiddleware.ValidateJwtToken, ipfsMiddleware.FetchFile)
			ipfs.Post("/upload", authMiddleware.ValidateJwtToken, ipfsMiddleware.UploadFile)
		}
	}

	app.Listen(":4000")
}

func loadKey() []byte {
	key := make([]byte, 24)
	encodedKey, err := ioutil.ReadFile(cutils.GetKeyPath())
	if err != nil {
		return cutils.GenerateKeyFile()
	}
	hex.Decode(key, encodedKey)
	return key
}

func ping(c *fiber.Ctx) error {
	return c.JSON(map[string]interface{}{
		"code":    200,
		"status":  "server is running",
		"message": c.Locals("vals"),
	})
}

func securedEndpoint(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"email": c.Locals("email"),
		// "user_uid": c.Locals("userUid"),
	})
}

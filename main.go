package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

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

	jwtUri := os.Getenv("JWT_VALIDATION_URI")
	adminHydraHost := os.Getenv("ADMIN_HYDRA_HOST")
	ipfsApiServer := os.Getenv("IPFS_API_SERVER_URI")
	ipfsGateway := os.Getenv("IPFS_GATEWAY_URI")

	fmt.Println("JWT VALIDATION URI = ", jwtUri)
	fmt.Println("ADMIN HYDRA HOST = ", adminHydraHost)
	fmt.Println("IPFS API SERVER = ", ipfsApiServer)
	fmt.Println("IPFS GATEWAY = ", ipfsGateway)

	cryptoService := services.NewCryptoService(loadKey())
	authService := services.NewAuthService(jwtUri, adminHydraHost)
	ipfsClient := ipfs.NewClient(ipfsApiServer, ipfsGateway)

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

		// Testing endpoint
		v1.Get("/secure", authMiddleware.ValidateJwtToken, securedEndpoint)
		v1.Get("/secureOauth", authMiddleware.IntrospectAccessToken, securedEndpoint)

		user := v1.Group("/user")
		{
			user.Get("/fetch", authMiddleware.ValidateJwtToken, ipfsMiddleware.FetchFile)
			user.Post("/upload", authMiddleware.ValidateJwtToken, ipfsMiddleware.UploadFile)
		}

		bank := v1.Group("/bank")
		{
			bank.Get("/fetch", authMiddleware.IntrospectAccessToken, ipfsMiddleware.FetchFile)
		}

	}

	// app.Listen(":4000")
	//Start server
	var port string
	if os.Getenv("PORT_LISTEN") != "" {
		port = fmt.Sprintf(":%s", os.Getenv("PORT_LISTEN"))
	}
	app.Listen(port)
}

func loadKey() []byte {
	fmt.Println("Loading key...")

	key := make([]byte, 24)
	encodedKey, err := ioutil.ReadFile(cutils.GetKeyPath())
	if err != nil {
		// Key is not found
		fmt.Println("Cannot find master key")
		// Generating new key file
		fmt.Println("Generating new master key file...")
		key = cutils.GenerateKeyFile()
		fmt.Println("Key is generated and stored in ", cutils.GetKeyDirPath())
		return key
	}
	fmt.Println("Key found, using this key as m master key")
	hex.Decode(key, encodedKey) // Decode key from hex to bytes
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
	})
}

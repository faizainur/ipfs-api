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
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://dashboard.catena.id, https://api.catena.id, https://catena.id, http://localhost:8080",
		AllowCredentials: true,
		AllowHeaders:     "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
		AllowMethods:     "GET, PUT, POST, OPTIONS",
	}))
	app.Use(logger.New())
	// app.Use(middleware)

	ipfsApiServer := os.Getenv("IPFS_API_SERVER_URI")
	ipfsGateway := os.Getenv("IPFS_GATEWAY_URI")
	mongoDbUri := "mongodb+srv://imblock:imblock@dev-catena.yuofs.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	// mongoDbUri := os.Getenv("MONGODB_URI")

	fmt.Println("IPFS API SERVER = ", ipfsApiServer)
	fmt.Println("IPFS GATEWAY = ", ipfsGateway)
	fmt.Println("Mongo DB Uri = ", mongoDbUri)

	cryptoService := services.NewCryptoService(loadKey())
	ipfsClient := ipfs.NewClient(ipfsApiServer, ipfsGateway)

	ipfsMiddleware := middlewares.IpfsMiddleware{
		IpfsClient:    ipfsClient,
		CryptoService: cryptoService,
	}


	v1 := app.Group("/api/v1/ipfs")
	{
		v1.Get("/ping", ping)

		// Testing endpoint

		user := v1.Group("/user")
		{
			user.Get("/fetch",  ipfsMiddleware.FetchFile)
			// user.Post("/upload", authMiddleware.ValidateJwtToken, ipfsMiddleware.UploadFile)
			user.Post("/upload",  ipfsMiddleware.UploadFile)
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
		"status":  "running",
		"message": "IPFS API is running",
	})
}

// func securedEndpoint(c *fiber.Ctx) error {
// 	return c.Status(200).JSON(fiber.Map{
// 		"email": c.Locals("email"),
// 	})
// }

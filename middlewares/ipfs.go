package middlewares

import (
	"bytes"
	"fmt"

	ipfs "github.com/faizainur/ipfs-api/ipfs_client"
	"github.com/faizainur/ipfs-api/services"
	"github.com/gofiber/fiber/v2"
)

type IpfsMiddleware struct {
	IpfsClient    *ipfs.IPFSClient
	CryptoService *services.CryptoService
}

func (f *IpfsMiddleware) UploadFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	files := form.File["file"]
	dataBuffer := new(bytes.Buffer)
	filename := files[0].Filename

	for _, file := range files {
		fh, err := file.Open()
		if err != nil {
			panic(err)
		}
		dataBuffer.ReadFrom(fh)
	}

	encryptedFile := f.CryptoService.AESEncrypt(dataBuffer.Bytes())

	resp, errUpload := f.IpfsClient.UploadFile(filename, encryptedFile)
	if errUpload != nil {
		return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("%s: %s", "Error", errUpload.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (f *IpfsMiddleware) FetchFile(c *fiber.Ctx) error {
	cid := c.Query("cid")

	data, err := f.IpfsClient.FetchFile(cid)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	decryptedFile := f.CryptoService.AESDecrypt(data)
	return c.Status(fiber.StatusOK).Send(decryptedFile)
}

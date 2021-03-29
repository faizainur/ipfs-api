package middlewares

import (
	"bytes"
	"fmt"
	"time"

	ipfs "github.com/faizainur/ipfs-api/ipfs_client"
	"github.com/faizainur/ipfs-api/services"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

type IpfsMiddleware struct {
	IpfsClient    *ipfs.IPFSClient
	CryptoService *services.CryptoService
}

func (f *IpfsMiddleware) UploadFile(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
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

	encryptedFile, err := f.CryptoService.EncryptUserFile(email, dataBuffer.Bytes())
	if err != nil {
		c.SendString(err.Error())
	}

	resp, errUpload := f.IpfsClient.UploadFile(filename, encryptedFile)
	if errUpload != nil {
		return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("%s: %s", "Error", errUpload.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (f *IpfsMiddleware) FetchFile(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	cid := c.Query("cid")

	data, err := f.IpfsClient.FetchFile(cid)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	decryptedFile := f.CryptoService.DecryptUserFile(email, data)
	typeFile := mimetype.Detect(decryptedFile)

	fileName := fmt.Sprintf("%d%s", time.Now().Unix(), typeFile.Extension())
	c.Attachment(fileName)
	// c.Set("Content-Disposition", "inline")
	c.Set("Content-Type", typeFile.String())
	return c.Status(fiber.StatusOK).Send(decryptedFile)
}

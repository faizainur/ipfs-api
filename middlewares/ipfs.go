package middlewares

import (
	"bytes"
	"fmt"

	ipfs "github.com/faizainur/ipfs-api/ipfs_client"
	"github.com/gofiber/fiber/v2"
)

type IpfsMiddleware struct {
	IpfsClient *ipfs.IPFSClient
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

	resp, errUpload := f.IpfsClient.UploadFile(filename, dataBuffer.Bytes())
	if errUpload != nil {
		return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("%s: %s", "Error", errUpload.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

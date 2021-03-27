package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

func testPos(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "Server is running",
	})
}

type ipfsResponse struct {
	Name string `json:"name,omitempty"  bson:"name"  form:"name"  binding:"name"`
	Hash string `json:"hash,omitempty"  bson:"hash"  form:"hash"  binding:"hash"`
	Size string `json:"size,omitempty"  bson:"size"  form:"size"  binding:"size"`
}

// var client = ipfs.IPFSClient{
// 	ApiServerUri:     "http://localhost:5001/api/v0/",
// 	GatewayServerUri: "http://localhost:7000/ipfs/",
// }

func Upload(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		log.Fatal(err.Error())
	}

	buffer := multipart.Form{
		Value: form.Value,
		File:  form.File,
	}

	files := buffer.File["file"]
	buf := new(bytes.Buffer)

	for _, file := range files {

		fh, err := file.Open()
		if err != nil {
			log.Fatal(err.Error())
		}
		buf.ReadFrom(fh)
	}

	fmt.Print(buf.Bytes(), "\n")

	data, errUploadIpfs := uploadIpfs(buf.Bytes())
	if errUploadIpfs != nil {
		panic(errUploadIpfs)
	}

	var ipfsResponse ipfsResponse

	json.Unmarshal(data, &ipfsResponse)

	fmt.Print(string(data))
	return c.Send(buf.Bytes())

}

func uploadIpfs(data []byte) ([]byte, error) {

	// fmt.Print(data)

	file := fiber.AcquireFormFile()
	file.Content = data
	file.Fieldname = "ipfs test"
	file.Name = "IPFSTest.txt"

	agent := fiber.AcquireAgent()
	resp := fiber.AcquireResponse()

	defer func() {
		fiber.ReleaseFormFile(file)
		fiber.ReleaseResponse(resp)
		fiber.ReleaseAgent(agent)
	}()

	req := agent.Request()
	req.Header.SetMethod(fiber.MethodPost)
	req.SetRequestURI("http://127.0.0.1:5001/api/v0/add")

	agent.UserAgent("IPFS API Server")
	// agent.SendFile("./test.jpg", "test ipfs").MultipartForm(nil)
	agent.FileData(file).MultipartForm(nil)

	if err := agent.Parse(); err != nil {
		return nil, err
	}

	if err := agent.HostClient.Do(req, resp); err != nil {
		return nil, err
	}

	return resp.Body(), nil
}

func DownloadData(c *fiber.Ctx) error {
	// client := fiber.AcquireClient()
	agent := fiber.AcquireAgent()
	resp := fiber.AcquireResponse()

	defer fiber.ReleaseResponse(resp)
	defer fiber.ReleaseAgent(agent)

	req := agent.Request()
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI("http://127.0.0.1:7000/ipfs/QmRoKkkuQ8WLWVFwX6htDHED7Hm52TpgVincoFAL5YBoiK")

	agent.UserAgent("IPFS API Server")
	agent.JSONDecoder(json.Unmarshal)
	agent.JSONEncoder(json.Marshal)
	// agent.SetResponse(resp)

	if err := agent.Parse(); err != nil {
		panic(err)
	}
	// fmt.Print("H", resp.Body())
	// agent = client.Get("http://127.0.0.1:7000/ipfs/QmRoKkkuQ8WLWVFwX6htDHED7Hm52TpgVincoFAL5YBoiK").SetResponse(resp)
	// agent.SetResponse(resp)

	// _, data, _ := agent.Bytes()

	err := agent.HostClient.Do(req, resp)
	if err != nil {
		log.Fatal(err.Error())
	}

	out, err := os.Create("test.jpg")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer out.Close()

	r := io.NewSectionReader(bytes.NewReader(resp.Body()), 0, int64(len(resp.Body())))
	r2 := io.NewSectionReader(bytes.NewReader(resp.Body()), 0, int64(len(resp.Body())))

	contentType, errMime := mimetype.DetectReader(r2)
	if errMime != nil {
		fmt.Println("Error")
		log.Fatal(errMime.Error())
	}
	fmt.Println("Hello", contentType.String(), contentType.Extension())

	_, errWrite := io.Copy(out, r)
	if errWrite != nil {
		log.Fatal(errWrite.Error())
	}

	// fmt.Print(data)
	// c.SendFile(/)
	// fmt.Print(resp.Body())
	return c.Status(fiber.StatusOK).Send(resp.Body())
}

func isAdmin(c *fiber.Ctx) error {
	isAdmin, err := strconv.ParseBool(c.FormValue("is_admin"))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"code":  http.StatusInternalServerError,
			"error": err.Error(),
		})
	}

	if !isAdmin {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"code":  http.StatusUnauthorized,
			"error": "Unauthorized access",
		})
	}

	return c.Next()
}

func middleware(c *fiber.Ctx) error {
	c.Locals("vals", "this is a value from middleware")
	return c.Next()
}

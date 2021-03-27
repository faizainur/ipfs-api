package ipfs

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	AddFileEndpoint = "add"
	CatFileEndpoint = "cat"
)

type ipfsUploadResponse struct {
	Name string `json:"name,omitempty"  bson:"name"  form:"name"  binding:"name"`
	Hash string `json:"hash,omitempty"  bson:"hash"  form:"hash"  binding:"hash"`
	Size string `json:"size,omitempty"  bson:"size"  form:"size"  binding:"size"`
}

type IPFSClient struct {
	ApiServerUri     string
	GatewayServerUri string
}

func (f *IPFSClient) FetchFile(cid string) ([]byte, error) {
	uri := f.formFetchUri(cid)

	agent := fiber.AcquireAgent()
	resp := fiber.AcquireResponse()

	defer func() {
		fiber.ReleaseResponse(resp)
		fiber.ReleaseAgent(agent)
	}()

	agent.UserAgent("IPFS API Server")

	req := agent.Request()
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI(uri)

	if err := agent.Parse(); err != nil {
		panic(err)
	}

	if err := agent.HostClient.Do(req, resp); err != nil {
		return nil, err
	}

	return resp.Body(), nil
}

func (f *IPFSClient) UploadFile(filename string, data []byte) (ipfsUploadResponse, error) {
	var jsonResponse ipfsUploadResponse

	file := fiber.AcquireFormFile()
	file.Content = data
	file.Name = filename

	agent := fiber.AcquireAgent()
	resp := fiber.AcquireResponse()

	defer func() {
		fiber.ReleaseFormFile(file)
		fiber.ReleaseResponse(resp)
		fiber.ReleaseAgent(agent)
	}()

	req := agent.Request()
	req.Header.SetMethod(fiber.MethodPost)
	req.SetRequestURI(f.formApiIpfsUri(AddFileEndpoint, nil))

	agent.UserAgent("IPFS API Server")
	agent.FileData(file).MultipartForm(nil)

	if err := agent.Parse(); err != nil {
		panic(err)
	}

	if err := agent.HostClient.Do(req, resp); err != nil {
		return ipfsUploadResponse{}, err
	}

	json.Unmarshal(resp.Body(), &jsonResponse)

	return jsonResponse, nil
}

func (f *IPFSClient) formFetchUri(cid string) string {
	var builder strings.Builder

	builder.WriteString(f.GatewayServerUri)
	builder.WriteString(cid)

	return builder.String()
}

func (f *IPFSClient) formApiIpfsUri(endpoint string, queryString map[string]string) string {
	var builder strings.Builder

	builder.WriteString(f.ApiServerUri)
	builder.WriteString(endpoint)

	if queryString != nil {
		builder.WriteString("?")
		for key, val := range queryString {
			builder.WriteString(key)
			builder.WriteString("=")
			builder.WriteString(val)
			builder.WriteString("&")
		}
	}

	return builder.String()
}

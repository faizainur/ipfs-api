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

// IPFSClient represent the object of IFPS client
type IPFSClient struct {
	apiServerUri     string // IPFS API Server Uri
	gatewayServerUri string // IPFS HTTP Gateway
}

// NewClient creates and returns an object of IPFSClient
// Takes apiServerUri for the URI of IPFS Server
// And gatewayServerUri for IPFS HTTP Gateway URI
func NewClient(apiServerUri string, gatewayServerUri string) *IPFSClient {
	return &IPFSClient{
		apiServerUri:     apiServerUri,
		gatewayServerUri: gatewayServerUri,
	}
}

// FetchFile will fetch a file stored in IPFS network
// Find a file based on the given CID of the file
// returns the file in bytes
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

// UploadFile upload a gile to IPFS network
// This function takes two parameters, filename in string and file data in bytes
// File is passed to IPFS API endpoint using HTTP form data
// return an object of IpfsUploadResponse
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

// formFetchUri will create a Fetch Uri for the given cid
// return URI as string
func (f *IPFSClient) formFetchUri(cid string) string {
	var builder strings.Builder
	builder.WriteString(f.gatewayServerUri)
	builder.WriteString(cid)
	return builder.String()
}

// formIpfsUri will create a API IPFS endpoint uri
// return URI as string
func (f *IPFSClient) formApiIpfsUri(endpoint string, queryString map[string]string) string {
	var builder strings.Builder

	builder.WriteString(f.apiServerUri)
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

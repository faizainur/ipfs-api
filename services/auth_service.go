package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/gofiber/fiber/v2"
	hydraClient "github.com/ory/hydra-client-go/client"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
)

type JwtTokenValidationData struct {
	Email   string `json:"email,omitempty"  bson:"email"  form:"email"  binding:"email"`
	UserUid string `json:"user_uid,omitempty"  bson:"user_uid"  form:"user_uid"  binding:"user_uid"`
}

type JwtTokenValidationResponse struct {
	Code    int                    `json:"code,omitempty"  bson:"code"  form:"code"  binding:"code"`
	Data    JwtTokenValidationData `json:"data,omitempty"  bson:"data"  form:"data"  binding:"data"`
	IsValid bool                   `json:"is_valid,omitempty"  bson:"is_valid"  form:"is_valid"  binding:"is_valid"`
}

type AuthService struct {
	jwtValidationUri string
	hydraAdmin       admin.ClientService
}

func NewAuthService(jwtValidationUri string, hydraHost string) *AuthService {
	skipTlsClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 10 * time.Second,
	}
	transport := httptransport.NewWithClient(hydraHost, "/", []string{"https"}, skipTlsClient)
	hydra := hydraClient.New(transport, nil)

	return &AuthService{
		jwtValidationUri: jwtValidationUri,
		hydraAdmin:       hydra.Admin,
	}
}

func (a *AuthService) ValidateJwt(token string) (bool, JwtTokenValidationData, error) {

	var jsonResponse JwtTokenValidationResponse

	agent := fiber.AcquireAgent()
	resp := fiber.AcquireResponse()

	defer func() {
		fiber.ReleaseResponse(resp)
		fiber.ReleaseAgent(agent)
	}()

	agent.UserAgent("IPFS API Server")

	authToken := fmt.Sprintf("Bearer %s", token)

	req := agent.Request()
	req.Header.SetMethod(fiber.MethodPost)
	req.Header.Add("Authorization", authToken)
	req.SetRequestURI(a.jwtValidationUri)

	if err := agent.Parse(); err != nil {
		fmt.Println("error parse host client")

		panic(err)
	}

	if err := agent.HostClient.Do(req, resp); err != nil {
		fmt.Println("error http ")

		return false, JwtTokenValidationData{}, err
	}

	err := json.Unmarshal(resp.Body(), &jsonResponse)
	if err != nil {
		fmt.Println("Error json response ")

		return false, JwtTokenValidationData{}, nil
	}

	if !jsonResponse.IsValid {
		fmt.Println("Not valid ")

		return false, JwtTokenValidationData{}, nil
	}

	return true, jsonResponse.Data, nil
}

func (a *AuthService) IntrospectTokenOauth2(token string) (bool, *models.OAuth2TokenIntrospection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := admin.NewIntrospectOAuth2TokenParams()
	params.WithContext(ctx)
	params.SetToken(token)

	responseIntrospection, err := a.hydraAdmin.IntrospectOAuth2Token(params)
	if err != nil {
		return false, &models.OAuth2TokenIntrospection{}, err
	}

	isActive := *responseIntrospection.GetPayload().Active
	if !isActive {
		return false, &models.OAuth2TokenIntrospection{}, nil
	}
	return true, responseIntrospection.GetPayload(), nil
}

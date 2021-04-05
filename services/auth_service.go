package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/gofiber/fiber/v2"
	"github.com/ory/hydra-client-go/client"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
)

// JwtTokenValidatoinData is a model represent the user data
// Consists of email address and User UID
// Used for a return object after an authentication mechanism
type JwtTokenValidationData struct {
	Email   string `json:"email,omitempty"  bson:"email"  form:"email"  binding:"email"`
	UserUid string `json:"user_uid,omitempty"  bson:"user_uid"  form:"user_uid"  binding:"user_uid"`
}

// JwtTokenValidationesponse
type JwtTokenValidationResponse struct {
	Code    int                    `json:"code,omitempty"  bson:"code"  form:"code"  binding:"code"`
	Data    JwtTokenValidationData `json:"data,omitempty"  bson:"data"  form:"data"  binding:"data"`
	IsValid bool                   `json:"is_valid,omitempty"  bson:"is_valid"  form:"is_valid"  binding:"is_valid"`
}

// AuthService represent the object of authentication service
// It has methods use for performing authentication mechanism
// Supported mechanism : JWT Validation and Introspect Access Token for Oauth2
type AuthService struct {
	jwtValidationUri string              // IDP Endpoin for JWT validation
	hydraAdmin       admin.ClientService // Hydra Admin Object
}

// NewAuthService creates and return a new instance of AuthService object
// Takes jwtValidationUri and hydraAdminHost as parameters required
// for creating a new instances of AuthService object
func NewAuthService(jwtValidationUri string, hydraAdminHost string) *AuthService {
	var (
		hydraAdmin            *client.OryHydra
		enableTlsVerification string
	)

	// Get ENABLE_TLS_VERIFICATION environment variable
	enableTlsVerification = os.Getenv("ENABLE_TLS_VERIFICATION")

	// If environment variable is not set
	if enableTlsVerification == "" {
		// Set default value "0"
		// Default : Disable TLS verification
		enableTlsVerification = "0"
	}

	if enableTlsVerification == "0" {
		// If TLS verification is disabled
		// create a new HTTP client
		// Set TLS InsecureSkipVerify to true to skip TLS verification
		skipTlsClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 10 * time.Second, // Set HTTP timeout to 10 seconds
		}
		transport := httptransport.NewWithClient(hydraAdminHost, "/", []string{"https"}, skipTlsClient)
		hydraAdmin = client.New(transport, nil) // Create a new hydra admin client

	} else if enableTlsVerification == "1" {
		// Disclaimer : This feature is in BETA
		// If TLS verification is enabled
		// Create a new hydra admin client using the default function
		// Provided by Ory Hydra package
		adminUrl, _ := url.Parse(hydraAdminHost)
		hydraAdmin = client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
			Schemes:  []string{adminUrl.Scheme},
			Host:     adminUrl.Host,
			BasePath: adminUrl.Path,
		})
	}

	return &AuthService{
		jwtValidationUri: jwtValidationUri,
		hydraAdmin:       hydraAdmin.Admin,
	}
}

// ValidateJwt will validate the provided JWT token
// If a user is authenticated, this function will return
// the user data consists of user email address and user uid
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

// IntrospectTokenOauth2 will intraspect the given Oauth2 access token
// This method will send a request to Introspect Token endpoint
// In Oauth2 server
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

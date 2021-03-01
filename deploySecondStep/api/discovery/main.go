package main

import (
	"fmt"
	"log"
	"context"
	"net/http"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type APIResponse struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported"`
	UserinfoEndpoint                           string   `json:"userinfo_endpoint"`
	JwksUri                                    string   `json:"jwks_uri"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	SubjectTypesSupported                      []string `json:"subject_types_supported"`
	UserinfoSigningAlgValuesSupported          []string `json:"userinfo_signing_alg_values_supported"`
	IdTokenSigningAlgValuesSupported           []string `json:"id_token_signing_alg_values_supported"`
	RequestObjectSigningAlgValuesSupported     []string `json:"request_object_signing_alg_values_supported"`
	DisplayValuesSupported                     []string `json:"display_values_supported"`
	ClaimsSupported                            []string `json:"claims_supported"`
}

type APIErrorResponse struct {
	Message string `json:"message"`
}

type Response events.APIGatewayProxyResponse

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	res, err := openIdConfiguration(ctx, request.RequestContext.Stage, request.RequestContext.DomainName)
	if err == nil {
		jsonBytes, _ = json.Marshal(res)
	}
	log.Print(request.RequestContext.Identity.SourceIP)
	if err != nil {
		log.Print(err)
		jsonBytes, _ = json.Marshal(APIErrorResponse{Message: fmt.Sprint(err)})
		return Response{
			StatusCode: http.StatusInternalServerError,
			Body: string(jsonBytes),
		}, nil
	}
	return Response {
		StatusCode: http.StatusOK,
		Body: string(jsonBytes),
	}, nil
}

func openIdConfiguration(ctx context.Context, stage string, domainName string)(APIResponse, error) {
	var res APIResponse
	var host string
	if  len(stage) > 0 {
		host = domainName + "/" + stage
	} else {
		host = domainName
	}
	res.Issuer = "https://" + host
	res.AuthorizationEndpoint = "https://" + host + "/authorize"
	res.TokenEndpoint = "https://" + host + "/token"
	res.TokenEndpointAuthMethodsSupported = []string{"client_secret_basic", "private_key_jwt"}
	res.TokenEndpointAuthSigningAlgValuesSupported = []string{"RS256"}
	res.UserinfoEndpoint = "https://" + host + "/userinfo"
	res.JwksUri = "https://" + host + "/.well-known/jwks.json"
	res.ScopesSupported = []string{"openid", "read:user", "user:email"}
	res.ResponseTypesSupported = []string{"code", "code id_token", "id_token", "token id_token"}
	res.SubjectTypesSupported = []string{"public"}
	res.UserinfoSigningAlgValuesSupported = []string{"none"}
	res.IdTokenSigningAlgValuesSupported = []string{"RS256"}
	res.RequestObjectSigningAlgValuesSupported = []string{"none"}
	res.DisplayValuesSupported = []string{"page", "popup"}
	res.ClaimsSupported = []string{
		"sub",
		"name",
		"preferred_username",
		"profile",
		"picture",
		"website",
		"email",
		"email_verified",
		"updated_at",
		"iss",
		"aud",
	}
	return res, nil
}

func main() {
	lambda.Start(HandleRequest)
}

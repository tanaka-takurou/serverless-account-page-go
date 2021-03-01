package main

import (
	"os"
	"fmt"
	"log"
	"errors"
	"context"
	"net/url"
	"net/http"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type APIResponse struct {
	Message string `json:"message"`
}

type Response events.APIGatewayProxyResponse

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var url string
	var err error
	url, err = getAuthorizeUrl(request.QueryStringParameters)
	log.Print(request.RequestContext.Identity.SourceIP)
	if err != nil {
		log.Print(err)
		jsonBytes, _ = json.Marshal(APIResponse{Message: fmt.Sprint(err)})
		return Response{
			StatusCode: http.StatusInternalServerError,
			Body: string(jsonBytes),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}
	return Response {
		StatusCode: http.StatusFound,
		Body: string(jsonBytes),
		Headers: map[string]string{
			"Location": url,
		},
	}, nil
}

func getAuthorizeUrl(param map[string]string)(string, error) {
	if len(param["scope"]) < 1 || len(param["state"]) < 1 || len(param["response_type"]) < 1 {
		return "", errors.New("Insufficient parameters")
	}
	return os.Getenv("GITHUB_LOGIN_URL") + "/oauth/authorize?client_id=" + os.Getenv("GITHUB_CLIENT_ID") + "&scope=" + url.PathEscape(param["scope"]) +"&state=" + param["state"] + "&response_type=" + param["response_type"], nil
}

func main() {
	lambda.Start(HandleRequest)
}

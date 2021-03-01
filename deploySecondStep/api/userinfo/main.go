package main

import (
	"io"
	"os"
	"fmt"
	"log"
	"time"
	"bytes"
	"regexp"
	"context"
	"strings"
	"strconv"
	"net/http"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type APIResponse struct {
	Sub               string `json:"sub"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Profile           string `json:"profile"`
	Picture           string `json:"picture"`
	Website           string `json:"website"`
	UpdatedAt         string `json:"updated_at"`
	Email             string `json:"email"`
	EmailVerified     bool`json:"email_verified"`
}

type APIErrorResponse struct {
	Message string `json:"message"`
}

type Response events.APIGatewayProxyResponse

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	accessToken, err := getBearerToken(request)
	if err == nil {
		res, err := getUserInfo(accessToken)
		if err == nil {
			jsonBytes, _ = json.Marshal(res)
		}
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

func getBearerToken(request events.APIGatewayProxyRequest)(string, error) {
	if authHeader, ok := request.Headers["Authorization"]; ok {
		// Section 2.1 Authorization request header
		authHeaderValues := strings.Split(authHeader, " ")
		if len(authHeaderValues) > 1 {
			return authHeaderValues[1], nil
		}
	}

	if accessToken, ok := request.QueryStringParameters["access_token"]; ok {
		// Section 2.3 URI query parameter
		return accessToken, nil
	}

	if request.Headers["Content-Type"] == "application/x-www-form-urlencoded" {
		// Section 2.2 form encoded body parameter
		param := parseQueryString(request.Body)
		if len(param["access_token"]) > 1 {
			return param["access_token"], nil
		}
	}
	return "", fmt.Errorf("No token specified in request")
}

func getUserInfo(accessToken string)(APIResponse, error) {
	res := APIResponse{}

	userDetailsRaw, err := getUserDetails(accessToken)
	if err != nil {
		return res, err
	}
	userDetails := make(map[string]string)
	json.Unmarshal(userDetailsRaw, &userDetails)
	res.Sub = userDetails["id"]
	if len(res.Sub) < 1 {
		log.Print("No id")
		res.Sub = userDetails["login"]
	}
	res.Name = userDetails["name"]
	if len(res.Name) < 1 {
		log.Print("No name")
		res.Name = userDetails["login"]
	}
	res.PreferredUsername = userDetails["login"]
	res.Profile = userDetails["html_url"]
	res.Picture = userDetails["avatar_url"]
	res.Website = userDetails["blog"]
	res.UpdatedAt = convertDateToSecond(userDetails["updated_at"]) // OpenID requires the seconds since epoch in UTC

	userEmailsRaw, err := getUserEmails(accessToken)
	if err != nil {
		log.Print(err)
		return res, err
	}
	userEmails := make([]map[string]interface{}, 2)
	json.Unmarshal(userEmailsRaw, &userEmails)
	for _, v := range userEmails {
		if primaryRaw, ok := v["primary"]; ok {
			if primaryRaw.(bool) == true {
				res.Email = v["email"].(string)
				res.EmailVerified = v["verified"].(bool)
			}
		}
	}

	return res, nil
}

func getUserDetails(accessToken string)([]byte, error) {
	url := os.Getenv("GITHUB_API_URL") + "/user"
	return requestToGithub(url, accessToken)
}

func getUserEmails(accessToken string)([]byte, error) {
	url := os.Getenv("GITHUB_API_URL") + "/user/emails"
	return requestToGithub(url, accessToken)
}

func requestToGithub(url string, accessToken string)([]byte, error) {
	method := "GET"
	header := map[string]string{
		"Accept": "application/vnd.github.v3+json",
		"Content-Type": "application/json",
		"Authorization": "token " + accessToken,
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return []byte{}, err
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, res.Body)

	return buf.Bytes(), nil
}

func convertDateToSecond(date string) string {
	t, err := time.Parse("2006-01-02T15:04:05Z", date)
	if err != nil {
		log.Print(err)
		return "0"
	}
	return strconv.FormatInt(t.Unix(),10)
}

func parseQueryString(s string) map[string]string {
	res := map[string]string{}
	for _, v := range regexp.MustCompile("[&]").Split(s, -1) {
		w := regexp.MustCompile("[=]").Split(v, -1)
		if len(w) > 1 {
			res[w[0]] = w[1]
		}
	}
	return res
}

func main() {
	lambda.Start(HandleRequest)
}

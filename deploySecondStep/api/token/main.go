package main

import (
	"io"
	"os"
	"fmt"
	"log"
	"time"
	"bytes"
	"errors"
	"regexp"
	"context"
	"strings"
	"net/http"
	"crypto/rsa"
	"encoding/json"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type APIResponse struct {
	Scope       string `json:"scope"`
	IdToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

type APIErrorResponse struct {
	Message string `json:"message"`
}

type RequestParameter struct {
	GrantType    string `json:"grant_type"`
	RedirectUri  string `json:"redirect_uri"`
	ClientId     string `json:"client_id"`
	ResponseType string `json:"response_type"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	State        string `json:"state"`
}

type Response events.APIGatewayProxyResponse

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	var param map[string]string
	log.Print(request.HTTPMethod)
	if request.HTTPMethod == "GET" {
		param = request.QueryStringParameters
	} else {
		if request.Headers["Content-Type"] == "application/json" {
			param = parseJson(request.Body)
		} else {
			param = parseQueryString(request.Body)
		}
	}
	if len(param["code"]) < 1 {
		err = errors.New("Insufficient parameters")
	} else {
		res, err := getToken(param["code"], param["state"], request.RequestContext.Stage, request.RequestContext.DomainName)
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

func getToken(code string, state string, stage string, domainName string)(APIResponse, error) {
	res := APIResponse{}
	var host string
	if  len(stage) > 0 {
		host = domainName + "/" + stage
	} else {
		host = domainName
	}

	githubTokenRaw, err := requestToGithub(code, state)
	if err != nil {
		log.Print(err);
		return res, err
	}

	githubToken := make(map[string]string)
	json.Unmarshal(githubTokenRaw, &githubToken)
	if scopeRaw, ok := githubToken["scope"]; ok {
		res.TokenType = githubToken["token_type"]
		res.AccessToken = githubToken["access_token"]
		res.Scope = "openid " + strings.Replace(scopeRaw, ",", " ", -1)
	} else {
		log.Print(string(githubTokenRaw));
		err = errors.New("Failed to get access_token")
		return res, err
	}

	key, err := os.ReadFile(os.Getenv("CERT_PATH"))
	if err != nil {
		log.Print(err);
		return res, err
	}
	var signKey *rsa.PrivateKey
	if signKey, err = jwt.ParseRSAPrivateKeyFromPEM(key); err != nil {
		return res, err
	}

	token := jwt.New(jwt.SigningMethodRS256)

	token.Header["alg"] = os.Getenv("ALGORITHM")
	token.Header["kid"] = os.Getenv("KEY_ID")
	claims := token.Claims.(jwt.MapClaims)
	claims["aud"] = os.Getenv("GITHUB_CLIENT_ID")
	claims["iss"] = "https://" + host
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	tokenString, err := token.SignedString(signKey)
	if err != nil {
		return res, err
	}
	res.IdToken = tokenString
	return res, nil
}

func requestToGithub(code string, state string)([]byte, error) {
	param := RequestParameter{}
	param.GrantType = "authorization_code"
	param.RedirectUri = os.Getenv("COGNITO_REDIRECT_URI")
	param.ClientId = os.Getenv("GITHUB_CLIENT_ID")
	param.ResponseType = "code"
	param.ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	param.Code = code
	param.State = state

	requestJson, err := json.Marshal(param)
	if err != nil {
		return []byte{}, err
	}

	url := os.Getenv("GITHUB_LOGIN_URL") + "/oauth/access_token"
	method := "POST"
	header := map[string]string{
		"Accept": "application/json",
		"Content-Type": "application/json",
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestJson))
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

func parseJson(s string) map[string]string {
	res := make(map[string]string)
	json.Unmarshal([]byte(s), &res)
	return res
}

func main() {
	lambda.Start(HandleRequest)
}

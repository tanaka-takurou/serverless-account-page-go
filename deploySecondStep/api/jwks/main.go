package main

import (
	"os"
	"fmt"
	"log"
	"context"
	"strings"
	"net/http"
	"encoding/json"
	"encoding/base64"
	"encoding/binary"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type JSONWebKey struct {
	Algorithm    string `json:"alg"`
	KeyId        string `json:"kid"`
	KeyType      string `json:"kty"`
	PublicKeyUse string `json:"use"`
	RSAModulus   string `json:"n"`
	RSAExponent  string `json:"e"`
}

type APIResponse struct {
	Keys []JSONWebKey `json:"keys"`
}

type APIErrorResponse struct {
	Message string `json:"message"`
}

type Response events.APIGatewayProxyResponse

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	res, err := getJwks()
	if err == nil {
		jsonBytes, _ = json.Marshal(APIResponse{Keys: []JSONWebKey{res}})
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

func getJwks()(JSONWebKey, error) {
	res := JSONWebKey{}
	res.Algorithm = os.Getenv("ALGORITHM")
	res.KeyId = os.Getenv("KEY_ID")
	res.KeyType = "RSA"
	res.PublicKeyUse = "sig"
	key, err := os.ReadFile(os.Getenv("PUB_KEY_PATH"))
	if err != nil {
		return res, err
	}
	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(key)
	if err != nil {
		return res, err
	}
	res.RSAModulus = joseBase64UrlEncode(verifyKey.N.Bytes())
	res.RSAExponent = joseBase64UrlEncode(serializeRSAPublicExponentParam(verifyKey.E))
	return res, nil
}

func joseBase64UrlEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func serializeRSAPublicExponentParam(e int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(e))
	var i int
	for i = 0; i < 8; i++ {
		if buf[i] != 0 {
			break
		}
	}
	return buf[i:]
}

func main() {
	lambda.Start(HandleRequest)
}

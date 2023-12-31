package main

import (
	"os"
	"io"
	"log"
	"bytes"
	"embed"
	"regexp"
	"context"
	"html/template"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type PageData struct {
	Title   string
	ClientId string
	CognitoUrl string
}

type Response events.APIGatewayProxyResponse

//go:embed templates
var templateFS embed.FS

const title string = "Sample Cognito OpenId Page"

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	tmp := template.New("tmp")
	var dat PageData
	p := request.PathParameters
	funcMap := template.FuncMap{
		"safehtml": func(text string) template.HTML { return template.HTML(text) },
	}
	buf := new(bytes.Buffer)
	fw := io.Writer(buf)
	dat.Title = title
	dat.ClientId = os.Getenv("CLIENT_ID")
	dat.CognitoUrl = os.Getenv("COGNITO_URL")
	if extractParameters(p["proxy"]) == "profile" {
		tmp = template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/profile.html", "templates/view.html", "templates/header.html"))
	} else {
		tmp = template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/index.html", "templates/view.html", "templates/header.html"))
	}
	if e := tmp.ExecuteTemplate(fw, "base", dat); e != nil {
		log.Fatal(e)
	}
	res := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            string(buf.Bytes()),
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
	}
	return res, nil
}

func extractParameters(proxyPathParameter string) string {
	if len(proxyPathParameter) > 0 {
		res := regexp.MustCompile("[/]").Split(proxyPathParameter, -1)
		if len(res) > 0 {
			return res[0]
		}
	}
	return ""
}

func main() {
	lambda.Start(HandleRequest)
}

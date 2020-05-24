package main

import (
	"fmt"
	"log"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type UserResponse struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type ErrorResponse struct {
	Message  string `json:"message"`
}

type Response events.APIGatewayProxyResponse

const clientId         string = "your_clientId"
const userPoolId       string = "your_userPoolId"

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	d := make(map[string]string)
	json.Unmarshal([]byte(request.Body), &d)
	if v, ok := d["action"]; ok {
		switch v {
		case "login" :
			if n, ok := d["name"]; ok {
				if p, ok := d["pass"]; ok {
					t, e := login(n, p)
					if e == nil {
						jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: t})
					} else {
						err = e
					}
				}
			}
		case "getuser" :
			if t, ok := d["token"]; ok {
				n, e := getuser(t)
				if e == nil {
					jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: ""})
				} else {
					err = e
				}
			}
		case "changepass" :
			if t, ok := d["token"]; ok {
				if p, ok := d["pass"]; ok {
					if np, ok := d["newpass"]; ok {
						err = changePass(t, p, np)
						jsonBytes, _ = json.Marshal(UserResponse{Name: "", Token: ""})
					}
				}
			}
		case "logout" :
			if t, ok := d["token"]; ok {
				err = logout(t)
				if err == nil {
					jsonBytes, _ = json.Marshal(UserResponse{Name: "ok", Token: "Logout"})
				}
			}
		case "signup" :
			if n, ok := d["name"]; ok {
				if p, ok := d["pass"]; ok {
					if m, ok := d["mail"]; ok {
						err = signup(n, p, m)
						jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: ""})
					}
				}
			}
		case "confirmsignup" :
			if n, ok := d["name"]; ok {
				if c, ok := d["code"]; ok {
					err = confirmSignup(n, c)
					jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: ""})
				}
			}
		}
	}
	log.Print(request.RequestContext.Identity.SourceIP)
	if err != nil {
		jsonBytes, _ = json.Marshal(ErrorResponse{Message: fmt.Sprint(err)})
		return Response{
			StatusCode: 500,
			Body: string(jsonBytes),
		}, nil
	}
	responseBody := ""
	if len(jsonBytes) > 0 {
		responseBody = string(jsonBytes)
	}
	return Response {
		StatusCode: 200,
		Body: responseBody,
	}, nil
}

func login(name string, pass string)(string, error) {
	svc := cognitoidentityprovider.New(session.New(), &aws.Config{
		Region: aws.String("ap-northeast-1"),
	})

	params := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(name),
			"PASSWORD": aws.String(pass),
		},
		ClientId: aws.String(clientId),
	}

	res, err := svc.InitiateAuth(params)
	if err != nil {
		return "", err
	}
	return aws.StringValue(res.AuthenticationResult.AccessToken), nil
}

func getuser(token string)(string, error) {
	svc := cognitoidentityprovider.New(session.New(), &aws.Config{
		Region: aws.String("ap-northeast-1"),
	})

	params := &cognitoidentityprovider.GetUserInput{
		AccessToken: aws.String(token),
	}

	res, err := svc.GetUser(params)
	if err != nil {
		return "", err
	}
	return aws.StringValue(res.Username), nil
}

func changePass(token string, pass string, newPass string) error {
	svc := cognitoidentityprovider.New(session.New(), &aws.Config{
		Region: aws.String("ap-northeast-1"),
	})

	params := &cognitoidentityprovider.ChangePasswordInput{
		AccessToken:      aws.String(token),
		PreviousPassword: aws.String(pass),
		ProposedPassword: aws.String(newPass),
	}

	_, err := svc.ChangePassword(params)
	if err != nil {
		return err
	}
	return nil
}

func logout(token string) error {
	svc := cognitoidentityprovider.New(session.New(), &aws.Config{
		Region: aws.String("ap-northeast-1"),
	})
	o_params := &cognitoidentityprovider.GlobalSignOutInput{
		AccessToken: aws.String(token),
	}
	_, err := svc.GlobalSignOut(o_params)
	if err != nil {
		return err
	}
	return nil
}

func signup(name string, pass string, mail string) error {
	svc := cognitoidentityprovider.New(session.New(), &aws.Config{
		Region: aws.String("ap-northeast-1"),
	})
	ua := &cognitoidentityprovider.AttributeType {
		Name: aws.String("email"),
		Value: aws.String(mail),
	}
	params := &cognitoidentityprovider.SignUpInput{
		Username: aws.String(name),
		Password: aws.String(pass),
		ClientId: aws.String(clientId),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			ua,
		},
	}

	_, err := svc.SignUp(params)
	if err != nil {
		return err
	}
	return nil
}

func confirmSignup(name string, confirmationCode string) error {
	svc := cognitoidentityprovider.New(session.New(), &aws.Config{
		Region: aws.String("ap-northeast-1"),
	})

	params := &cognitoidentityprovider.ConfirmSignUpInput{
		Username: aws.String(name),
		ConfirmationCode: aws.String(confirmationCode),
		ClientId: aws.String(clientId),
	}

	_, err := svc.ConfirmSignUp(params)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}

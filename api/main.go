package main

import (
	"os"
	"fmt"
	"log"
	"sort"
	"time"
	"bytes"
	"errors"
	"strings"
	"context"
	"encoding/json"
	"path/filepath"
	"encoding/base64"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type UserResponse struct {
	Name     string `json:"name"`
	Token    string `json:"token"`
	ImgUrl   string `json:"imgurl"`
}

type ErrorResponse struct {
	Message  string `json:"message"`
}

type ImgData struct {
	Img_Id  int    `dynamodbav:"img_id"`
	Name    string `dynamodbav:"name"`
	Status  int    `dynamodbav:"status"`
	Url     string `dynamodbav:"url"`
	Updated string `dynamodbav:"updated"`
}

type Response events.APIGatewayProxyResponse

var dynamodbClient *dynamodb.Client
var cognitoClient *cognitoidentityprovider.Client

const layout  string = "2006-01-02 15:04"
const layout2 string = "20060102150405"

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
					t, e := login(ctx, n, p)
					if e == nil {
						jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: t, ImgUrl: ""})
					} else {
						err = e
					}
				}
			}
		case "getuser" :
			if t, ok := d["token"]; ok {
				n, e := getuser(ctx, t)
				if e == nil {
					jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: "", ImgUrl: ""})
				} else {
					err = e
				}
			}
		case "changepass" :
			if t, ok := d["token"]; ok {
				if p, ok := d["pass"]; ok {
					if np, ok := d["newpass"]; ok {
						err = changePass(ctx, t, p, np)
						jsonBytes, _ = json.Marshal(UserResponse{Name: "", Token: "", ImgUrl: ""})
					}
				}
			}
		case "logout" :
			if t, ok := d["token"]; ok {
				err = logout(ctx, t)
				if err == nil {
					jsonBytes, _ = json.Marshal(UserResponse{Name: "ok", Token: "Logout", ImgUrl: ""})
				}
			}
		case "signup" :
			if n, ok := d["name"]; ok {
				if p, ok := d["pass"]; ok {
					if m, ok := d["mail"]; ok {
						err = signup(ctx, n, p, m)
						jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: "", ImgUrl: ""})
					}
				}
			}
		case "confirmsignup" :
			if n, ok := d["name"]; ok {
				if c, ok := d["code"]; ok {
					err = confirmSignup(ctx, n, c)
					jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: "", ImgUrl: ""})
				}
			}
		case "getimg" :
			if t, ok := d["token"]; ok {
				n, e := getuser(ctx, t)
				if e == nil {
					imgUrl, _ := getImage(ctx, os.Getenv("IMG_TABLE_NAME"), n)
					jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: "", ImgUrl: imgUrl})
				} else {
					err = e
				}
			}
		case "uploadimg" :
			if t, ok := d["token"]; ok {
				n, e := getuser(ctx, t)
				if e == nil {
					if v, ok := d["filename"]; ok {
						if w, ok := d["filedata"]; ok {
							imgUrl, _ := uploadImage(ctx, os.Getenv("IMG_TABLE_NAME"), os.Getenv("BUCKET_NAME"), v, w, n)
							jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: "", ImgUrl: imgUrl})
						}
					}
				} else {
					err = e
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

func login(ctx context.Context, name string, pass string)(string, error) {
	if cognitoClient == nil {
		cognitoClient = cognitoidentityprovider.NewFromConfig(getConfig(ctx))
	}

	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: cognitotypes.AuthFlowTypeUserPasswordAuth,
		AuthParameters: map[string]string{
			"USERNAME": name,
			"PASSWORD": pass,
		},
		ClientId: aws.String(os.Getenv("CLIENT_ID")),
	}

	res, err := cognitoClient.InitiateAuth(ctx, input)
	if err != nil {
		return "", err
	}
	return aws.ToString(res.AuthenticationResult.AccessToken), nil
}

func getuser(ctx context.Context, token string)(string, error) {
	if cognitoClient == nil {
		cognitoClient = cognitoidentityprovider.NewFromConfig(getConfig(ctx))
	}

	input := &cognitoidentityprovider.GetUserInput{
		AccessToken: aws.String(token),
	}

	res, err := cognitoClient.GetUser(ctx, input)
	if err != nil {
		return "", err
	}
	return aws.ToString(res.Username), nil
}

func changePass(ctx context.Context, token string, pass string, newPass string) error {
	if cognitoClient == nil {
		cognitoClient = cognitoidentityprovider.NewFromConfig(getConfig(ctx))
	}

	input := &cognitoidentityprovider.ChangePasswordInput{
		AccessToken:      aws.String(token),
		PreviousPassword: aws.String(pass),
		ProposedPassword: aws.String(newPass),
	}

	_, err := cognitoClient.ChangePassword(ctx, input)
	return err
}

func logout(ctx context.Context, token string) error {
	if cognitoClient == nil {
		cognitoClient = cognitoidentityprovider.NewFromConfig(getConfig(ctx))
	}

	input := &cognitoidentityprovider.GlobalSignOutInput{
		AccessToken: aws.String(token),
	}

	_, err := cognitoClient.GlobalSignOut(ctx, input)
	return err
}

func signup(ctx context.Context, name string, pass string, mail string) error {
	if cognitoClient == nil {
		cognitoClient = cognitoidentityprovider.NewFromConfig(getConfig(ctx))
	}

	ua := &cognitotypes.AttributeType {
		Name: aws.String("email"),
		Value: aws.String(mail),
	}
	input := &cognitoidentityprovider.SignUpInput{
		Username: aws.String(name),
		Password: aws.String(pass),
		ClientId: aws.String(os.Getenv("CLIENT_ID")),
		UserAttributes: []cognitotypes.AttributeType{
			*ua,
		},
	}

	_, err := cognitoClient.SignUp(ctx, input)
	return err
}

func confirmSignup(ctx context.Context, name string, confirmationCode string) error {
	if cognitoClient == nil {
		cognitoClient = cognitoidentityprovider.NewFromConfig(getConfig(ctx))
	}

	input := &cognitoidentityprovider.ConfirmSignUpInput{
		Username: aws.String(name),
		ConfirmationCode: aws.String(confirmationCode),
		ClientId: aws.String(os.Getenv("CLIENT_ID")),
	}

	_, err := cognitoClient.ConfirmSignUp(ctx, input)
	return err
}

func scan(ctx context.Context, tableName string, filt expression.ConditionBuilder, proj expression.ProjectionBuilder)(*dynamodb.ScanOutput, error)  {
	if dynamodbClient == nil {
		dynamodbClient = dynamodb.NewFromConfig(getConfig(ctx))
	}
	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		return nil, err
	}
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}
	res, err := dynamodbClient.Scan(ctx, input)
	return res, err
}

func put(ctx context.Context, tableName string, av map[string]dynamodbtypes.AttributeValue) error {
	if dynamodbClient == nil {
		dynamodbClient = dynamodb.NewFromConfig(getConfig(ctx))
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err := dynamodbClient.PutItem(ctx, input)
	return err
}

func update(ctx context.Context, tableName string, an map[string]string, av map[string]dynamodbtypes.AttributeValue, key map[string]dynamodbtypes.AttributeValue, updateExpression string) error {
	if dynamodbClient == nil {
		dynamodbClient = dynamodb.NewFromConfig(getConfig(ctx))
	}
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: an,
		ExpressionAttributeValues: av,
		TableName: aws.String(tableName),
		Key: key,
		ReturnValues:     dynamodbtypes.ReturnValueUpdatedNew,
		UpdateExpression: aws.String(updateExpression),
	}

	_, err := dynamodbClient.UpdateItem(ctx, input)
	return err
}

func getImgCount(ctx context.Context, imgTableName string)(int32, error)  {
	filt := expression.NotEqual(expression.Name("status"), expression.Value(-1))
	proj := expression.NamesList(expression.Name("img_id"), expression.Name("status"), expression.Name("url"), expression.Name("updated"))
	result, err := scan(ctx, imgTableName, filt, proj)
	if err != nil {
		return int32(0), err
	}
	return result.ScannedCount, nil
}

func putImg(ctx context.Context, imgTableName string, url string, name string) error {
	t := time.Now()
	count, err := getImgCount(ctx, imgTableName)
	if err != nil {
		return err
	}
	item := ImgData {
		Img_Id: int(count) + 1,
		Name: name,
		Status: 0,
		Url: url,
		Updated: t.Format(layout),
	}
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	err = put(ctx, imgTableName, av)
	if err != nil {
		return err
	}
	return nil
}

func getImage(ctx context.Context, imgTableName string, username string)(string, error) {
	filt := expression.Equal(expression.Name("name"), expression.Value(username))
	proj := expression.NamesList(expression.Name("img_id"), expression.Name("status"), expression.Name("url"), expression.Name("updated"))
	result, err := scan(ctx, imgTableName, filt, proj)
	if err != nil {
		log.Print(err)
		return "", err
	}
	var imgDataList []ImgData
	for _, i := range result.Items {
		item := ImgData{}
		err = attributevalue.UnmarshalMap(i, &item)
		if err != nil {
			log.Println(err)
		} else {
			imgDataList = append(imgDataList, item)
		}
	}
	if len(imgDataList) < 1 {
		log.Print("No Img")
		return "", nil
	}
	sort.Slice(imgDataList, func(i, j int) bool { return imgDataList[i].Img_Id > imgDataList[j].Img_Id })
	return imgDataList[0].Url, nil
}

func uploadImage(ctx context.Context, imgTableName string, bucketName string, filename string, filedata string, username string)(string, error) {
	t := time.Now()
	b64data := filedata[strings.IndexByte(filedata, ',')+1:]
	data, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		log.Print(err)
		return "", err
	}
	extension := filepath.Ext(filename)
	var contentType string

	switch extension {
	case ".jpg":
		contentType = "image/jpeg"
	case ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".png":
		contentType = "image/png"
	default:
		return "", errors.New("this extension is invalid")
	}
	filename_ := string([]rune(filename)[:(len(filename) - len(extension))]) + t.Format(layout2) + extension
	uploader := s3manager.NewUploader(s3.NewFromConfig(getConfig(ctx)))
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		ACL: s3types.ObjectCannedACLPublicRead,
		Bucket: aws.String(bucketName),
		Key: aws.String(filename_),
		Body: bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Print(err)
		return "", err
	}
	imgUrl := "https://" + bucketName + ".s3-" + os.Getenv("REGION") + ".amazonaws.com/" + filename_
	putImg(ctx, imgTableName, imgUrl, username)
	return imgUrl, nil
}

func getConfig(ctx context.Context) aws.Config {
	var err error
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		log.Print(err)
	}
	return cfg
}

func main() {
	lambda.Start(HandleRequest)
}

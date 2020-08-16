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
	"strconv"
	"context"
	"encoding/json"
	"path/filepath"
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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
	Img_Id  int    `json:"img_id"`
	Name    string `json:"name"`
	Status  int    `json:"status"`
	Url     string `json:"url"`
	Updated string `json:"updated"`
}

type Response events.APIGatewayProxyResponse

const layout       string = "2006-01-02 15:04"
const layout2      string = "20060102150405"

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
						jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: t, ImgUrl: ""})
					} else {
						err = e
					}
				}
			}
		case "getuser" :
			if t, ok := d["token"]; ok {
				n, e := getuser(t)
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
						err = changePass(t, p, np)
						jsonBytes, _ = json.Marshal(UserResponse{Name: "", Token: "", ImgUrl: ""})
					}
				}
			}
		case "logout" :
			if t, ok := d["token"]; ok {
				err = logout(t)
				if err == nil {
					jsonBytes, _ = json.Marshal(UserResponse{Name: "ok", Token: "Logout", ImgUrl: ""})
				}
			}
		case "signup" :
			if n, ok := d["name"]; ok {
				if p, ok := d["pass"]; ok {
					if m, ok := d["mail"]; ok {
						err = signup(n, p, m)
						jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: "", ImgUrl: ""})
					}
				}
			}
		case "confirmsignup" :
			if n, ok := d["name"]; ok {
				if c, ok := d["code"]; ok {
					err = confirmSignup(n, c)
					jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: "", ImgUrl: ""})
				}
			}
		case "getimg" :
			if t, ok := d["token"]; ok {
				n, e := getuser(t)
				if e == nil {
					imgUrl, _ := getImage(os.Getenv("IMG_TABLE_NAME"), n)
					jsonBytes, _ = json.Marshal(UserResponse{Name: n, Token: "", ImgUrl: imgUrl})
				} else {
					err = e
				}
			}
		case "uploadimg" :
			if t, ok := d["token"]; ok {
				n, e := getuser(t)
				if e == nil {
					if v, ok := d["filename"]; ok {
						if w, ok := d["filedata"]; ok {
							imgUrl, _ := uploadImage(os.Getenv("IMG_TABLE_NAME"), os.Getenv("BUCKET_NAME"), v, w, n)
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
		ClientId: aws.String(os.Getenv("CLIENT_ID")),
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
		ClientId: aws.String(os.Getenv("CLIENT_ID")),
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
		ClientId: aws.String(os.Getenv("CLIENT_ID")),
	}

	_, err := svc.ConfirmSignUp(params)
	if err != nil {
		return err
	}
	return nil
}

func scan(tableName string, filt expression.ConditionBuilder)(*dynamodb.ScanOutput, error)  {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, err
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}
	return svc.Scan(params)
}

func put(tableName string, av map[string]*dynamodb.AttributeValue) error {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err := svc.PutItem(input)
	return err
}

func get(tableName string, key map[string]*dynamodb.AttributeValue, att string)(*dynamodb.GetItemOutput, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: key,
		AttributesToGet: []*string{
			aws.String(att),
		},
		ConsistentRead: aws.Bool(true),
		ReturnConsumedCapacity: aws.String("NONE"),
	}
	return svc.GetItem(input)
}

func update(tableName string, an map[string]*string, av map[string]*dynamodb.AttributeValue, key map[string]*dynamodb.AttributeValue, updateExpression string) error {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: an,
		ExpressionAttributeValues: av,
		TableName: aws.String(tableName),
		Key: key,
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String(updateExpression),
	}

	_, err := svc.UpdateItem(input)
	return err
}

func getImgCount(imgTableName string)(*int64, error)  {
	result, err := scan(imgTableName, expression.NotEqual(expression.Name("status"), expression.Value(-1)))
	if err != nil {
		return nil, err
	}
	return result.ScannedCount, nil
}

func putImg(imgTableName string, url string, name string) error {
	t := time.Now()
	count, err := getImgCount(imgTableName)
	if err != nil {
		return err
	}
	item := ImgData {
		Img_Id: int(*count) + 1,
		Name: name,
		Status: 0,
		Url: url,
		Updated: t.Format(layout),
	}
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}
	err = put(imgTableName, av)
	if err != nil {
		return err
	}
	return nil
}

func updateImg(imgTableName string, img_id int, url string, updated string) error {
	an := map[string]*string{
		"#u": aws.String("url"),
		"#d": aws.String("updated"),
	}
	av := map[string]*dynamodb.AttributeValue{
		":u": {
			S: aws.String(url),
		},
		":d": {
			S: aws.String(updated),
		},
	}
	key := map[string]*dynamodb.AttributeValue{
		"img_id": {
			N: aws.String(strconv.Itoa(img_id)),
		},
	}
	updateExpression := "set #u = :u, #d = :d"

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: an,
		ExpressionAttributeValues: av,
		TableName: aws.String(imgTableName),
		Key: key,
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String(updateExpression),
	}

	_, err := svc.UpdateItem(input)
	if err != nil {
		return err
	}
	return nil
}

func getImage(imgTableName string, username string)(string, error) {
	result, err := scan(imgTableName, expression.Name("name").Equal(expression.Value(username)))
	if err != nil {
		log.Print(err)
		return "", err
	}
	var imgDataList []ImgData
	for _, i := range result.Items {
		item := ImgData{}
		err = dynamodbattribute.UnmarshalMap(i, &item)
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

func uploadImage(imgTableName string, bucketName string, filename string, filedata string, username string)(string, error) {
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
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION"))},
	)
	if err != nil {
		log.Print(err)
		return "", err
	}
	filename_ := string([]rune(filename)[:(len(filename) - len(extension))]) + t.Format(layout2) + extension
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		ACL: aws.String("public-read"),
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
	putImg(imgTableName, imgUrl, username)
	return imgUrl, nil
}

func main() {
	lambda.Start(HandleRequest)
}

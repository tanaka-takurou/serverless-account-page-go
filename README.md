# serverless-account-page kit
Simple kit for serverless account page using AWS Lambda.


## Dependence
- aws-lambda-go
- aws-sdk-go


## Requirements
- AWS (Lambda, API Gateway, Cognito)
- aws-cli
- golang environment


## Cognito Setting
 - Create User Pool
 - Open api/main.go and edit clientId, userPoolId.

## Usage

### Edit View
##### HTML
- Edit templates/index.html

##### CSS
- Edit static/css/main.css

##### Javascript
- Edit static/js/main.js

##### Image
- Add image file into static/img/
- Edit templates/header.html like as 'favicon.ico'.

### Deploy
Open scripts/deploy.sh and edit 'your_function_name'.

Open api/scripts/deploy.sh and edit 'your_api_function_name'.

Open constant/constant.json and edit 'your_api_url'.


Then run this command.

```
$ sh scripts/deploy.sh
$ cd api
$ sh scripts/deploy.sh
```

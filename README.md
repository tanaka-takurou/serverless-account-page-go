# serverless-account-page kit
Simple kit for serverless account page using AWS Lambda.
Related with GitHub App.


## Dependence
- aws-lambda-go
- jwt-go


## Requirements
- AWS (Lambda, API Gateway, Cognito)
- aws-sam-cli
- golang environment

## Reference
- github.com/TimothyJones/github-cognito-openid-wrapper


## Usage

## SetUp Step
- Deploy 1st Step
- Create a GitHub OAuth App ((instructions)[https://docs.github.com/en/developers/apps/creating-an-oauth-app], with the following settings:
  - Authorization callback URL: https://<Your Cognito Domain>/oauth2/idpresponse
  - Note down the Client ID and secret
- Deploy 2nd Step
- Configure the OIDC integration in AWS console for Cognito (described below, but following (these instructions)[https://docs.aws.amazon.com/ja_jp/cognito/latest/developerguide/cognito-user-pools-oidc-idp.html]).
- Deploy 3rd Step
- Edit App Client Callback-URL

### Deploy

##### 1st Step
```bash
cd deployFirstStep
AWS_PROFILE={profile} AWS_DEFAULT_REGION={region} DOMAIN_PREFIX={domain prefix} CALLBACK_URL={callback url} make bucket={bucket} stack={stack name} deploy
```

##### 2nd Step
```bash
cd deploySecondStep
ssh-keygen -t rsa -b 4096 -m PEM -f jwtRS256.key -N ''
openssl rsa -in jwtRS256.key -pubout -outform PEM -out jwtRS256.key.pub
chmod a+r jwtRS256.key
mv jwtRS256.key api/token
mv jwtRS256.key.pub api/jwks
make clean build
AWS_PROFILE={profile} AWS_DEFAULT_REGION={region} GITHUB_CLIENT_ID={GitHub Client Id} GITHUB_CLIENT_SECRET={GitHub Client Secret} COGNITO_REDIRECT_URI={https://<Your Cognito Domain>/oauth2/idpresponse} GITHUB_URL={GitHub Url} GITHUB_LOGIN_URL={GitHub Login Url} make bucket={bucket} stack={stack name} deploy
```

##### 3rd Step
```bash
cd deployThirdStep
AWS_PROFILE={profile} AWS_DEFAULT_REGION={region} COGNITO_CLIENT_ID={Cognito Client Id} COGNITO_URL={https://<Your Cognito Domain>} make bucket={bucket} stack={stack name} deploy
```

### Edit View
##### HTML
- Edit deployThirdStep/templates/index.html

##### CSS
- Edit deployThirdStep/static/css/main.css

##### Javascript
- Edit deployThirdStep/static/js/main.js

##### Image
- Add image file into deployThirdStep/static/img/
- Edit deployThirdStep/templates/header.html like as 'favicon.ico'.

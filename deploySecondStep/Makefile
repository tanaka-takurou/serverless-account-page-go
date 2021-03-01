root	:=		$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: clean build deploy

clean:
	$(MAKE) -C "${root}/api/discovery" clean
	$(MAKE) -C "${root}/api/authorize" clean
	$(MAKE) -C "${root}/api/token" clean
	$(MAKE) -C "${root}/api/userinfo" clean
	$(MAKE) -C "${root}/api/jwks" clean

build:
	$(MAKE) -C "${root}/api/discovery" build
	$(MAKE) -C "${root}/api/authorize" build
	$(MAKE) -C "${root}/api/token" build
	$(MAKE) -C "${root}/api/userinfo" build
	$(MAKE) -C "${root}/api/jwks" build

deploy:
	sam package --output-template-file "${root}"/packaged.yml --s3-bucket "${bucket}"
	sam deploy --stack-name "${stack}" --capabilities CAPABILITY_IAM --template-file "${root}/packaged.yml" --parameter-overrides GitHubClientId="${GITHUB_CLIENT_ID}" GitHubClientSecret="${GITHUB_CLIENT_SECRET}" CognitoRedirectUri="${COGNITO_REDIRECT_URI}" GitHubUrl="${GITHUB_URL}" GitHubLoginUrl="${GITHUB_LOGIN_URL}"

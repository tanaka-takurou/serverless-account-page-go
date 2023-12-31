root	:=		$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: deploy

deploy:
	sam package --output-template-file "${root}"/packaged.yml --s3-bucket "${bucket}"
	sam deploy --stack-name "${stack}" --capabilities CAPABILITY_IAM --template-file "${root}/packaged.yml" --parameter-overrides UserPoolDefaultDomainPrefixName="${DOMAIN_PREFIX}" UserPoolClientCallbackURL="${CALLBACK_URL}"

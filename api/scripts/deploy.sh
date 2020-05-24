#!/bin/bash
echo 'Updating API Lambda-Function...'
cd `dirname $0`/../
rm function.zip
rm main
GOOS=linux go build main.go
zip -g function.zip main
aws lambda update-function-code \
	--profile default \
	--function-name your_api_function_name \
	--zip-file fileb://`pwd`/function.zip \
	--cli-connect-timeout 6000 \
	--publish

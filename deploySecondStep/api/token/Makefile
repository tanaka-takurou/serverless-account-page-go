root	:=		$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: clean build

clean:
	rm -rfv bin

build:
	mkdir -p bin
	cp jwtRS256.key bin/
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/main

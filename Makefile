BINARY_NAME=kubectl-resource-quota
VERSION=v0.1.0
LDFLAGS=-ldflags "-X main.version=${VERSION}"

.PHONY: build clean install test

build:
	go build -v ${LDFLAGS} -o ${BINARY_NAME} .

build-all:
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BINARY_NAME}-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-windows-amd64.exe .

install: build
	cp ${BINARY_NAME} ${GOPATH}/bin/

clean:
	rm -f ${BINARY_NAME}*

test:
	go test -v ./...

lint:
	golangci-lint run

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

run:
	go run . --namespace default

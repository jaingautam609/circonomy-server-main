.PHONY: lint setup

setup:
	go get ./... && go mod verify && go mod tidy && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.1

lint:
	golangci-lint run --fix

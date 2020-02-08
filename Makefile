.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm -rf ./email/email

build:
	GOOS=linux GOARCH=amd64 go build -o email/email ./email

run-api:
	GOOS=linux GOARCH=amd64 go build -o email/email ./email
	sam local start-api

run:
	GOOS=linux GOARCH=amd64 go build -o email/email ./email
	sam local invoke -e events.json EmailFunction
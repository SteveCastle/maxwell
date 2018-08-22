build:
	go get github.com/aws/aws-lambda-go/lambda
	env GOOS=linux go build -ldflags="-s -w" -o bin/maxwell cmd/lambda/main.go
deploy: build
	serverless deploy
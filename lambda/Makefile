.PHONY=build deploy test clean deps
SRC=src/main.go src/request.go src/response.go
BIN=bin/lambda
ZIP=function.zip

deps:
	go get -u ./...

build: deps
	go fmt ./...
	#golint ./...
	GOOS=linux GOARCH=amd64 go build -o $(BIN) $(SRC)

deploy: build
	zip $(ZIP) $(BIN)
	aws lambda update-function-code --function-name $(ARN) --zip-file fileb://${ZIP}
	
test: deps
	go test -v ./... -cover

clean:
	rm -f $(BIN) $(ZIP)


.PHONY=build test clean
SRC=src/main.go
BIN=bin/lambda
ZIP=function.zip

build:
	go fmt ./...
	GOOS=linux go build -o $(BIN) $(SRC)
	zip $(ZIP) $(BIN)

test:
	go test -v ./...

clean:
	rm -f $(BIN) $(ZIP)

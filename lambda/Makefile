BINDIR:=./bin
BINARY:=bootstrap
ZIPFILE:=$(BINARY).zip
SRC=src/main.go src/request.go src/response.go
REPORT:=./report

build: deps
	mkdir -p $(BINDIR)
	GOARCH=amd64 GOOS=linux go build -tags lambda.norpc -o $(BINDIR)/$(BINARY) $(SRC)

deps:
	go get -u ./...

deploy: build
	(cd $(BINDIR) && zip -FS $(ZIPFILE) $(BINARY))
	aws lambda update-function-code --function-name $(ARN) --zip-file fileb://$(BINDIR)/$(ZIPFILE)
	
test: deps
	go test -v ./... -cover

clean:
	rm -f $(BIN) $(ZIP)


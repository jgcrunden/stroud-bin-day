package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler() (string, error) {
	return "Hello World", nil
}
func main() {
	lambda.Start(Handler)
}

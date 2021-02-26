package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"io"
	"net/http"
	"os"
	"strings"
)

func checkIfPostcodeIsInSDC(postcode string, postcodes []string) bool {
	result := false
	for _, v := range postcodes {
		if strings.HasPrefix(postcode, fmt.Sprintf("%s ", v)) {
			result = true
			break
		}
	}
	return result
}

func getUserPostcode(deviceID string, accessToken string, apiEndpoint string, client *http.Client) (result string, err error) {
	// /v1/devices/*deviceId*/settings/address/countryAndPostalCode
	url := fmt.Sprintf("%s/v1/devices/%s/settings/address/countryAndPostalCode", apiEndpoint, deviceID)
	req, err := http.NewRequest("GET", url, nil)
	bearer := "Bearer " + accessToken
	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	fmt.Println(string(body))
	result = "GL10 3GW"
	return
}
func handler(request Request) (Response, error) {
	//var response alexa.Response
	postcodes := strings.Split(os.Getenv("SDC_POSTCODES"), " ")
	fmt.Println(postcodes)
	deviceID := request.Context.System.Device.DeviceID
	accessToken := request.Context.System.APIAccessToken
	apiEndpoint := request.Context.System.APIEndpoint
	client := &http.Client{}
	result, err := getUserPostcode(deviceID, accessToken, apiEndpoint, client)
	if err != nil {
		// do something with error
	}
	fmt.Println(result)

	//switch request.Body.Intent.Name {
	//	case "GetBinDayInfoIntent":
	//   	response = HandleGetBinDayInfoIntent(request)
	//}
	fmt.Printf("Starting lambda\n")
	return NewSimpleResponse("Saying Hello", "Hello world!"), nil
}

func main() {
	lambda.Start(handler)
}

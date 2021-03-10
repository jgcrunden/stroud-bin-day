package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	tableName string
)

//Address a struct to store the postalCode when unmarshalling a request to the Alexa address endpoint
type Address struct {
	PostalCode string `json:"postalCode"`
}

// Item is a struct to hold dynamoDB entries
type Item struct {
	Postcode string
	UPRN     int64
}

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

// getUPRNFromDynamoDB uses postcode to query dyamoDB for UPRN. Returns URPN if entry exists, else returns -1
func getUPRNFromDynamoDB(postcode string, svc dynamodbiface.DynamoDBAPI) (UPRN int64) {
	UPRN = -1
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(
	return
}

// addPostcodeAndUPRNToDynamoDB adds the postcode and corresponding URPN to dynamoDB for future look up. Takes the postcode UPRN and returns true if successful, false if not
func addPostcodeAndUPRNToDynamoDB(postcode string, URPN int64) bool {
	result := false

	return result
}

// lookupUPRNForPostcodeViaAPI takes the postcode and calls external API to lookup UPRN. Returns UPRN if successful, -1 if not
func lookupUPRNForPostcodeViaAPI(postcode string, client *http.Client) (URPN int64) {
	return -1
}

// getMyHousePageFromStroudGov takes the URPN and an http Client and makes an http request to stroud.gov.uk website. Returns the html page containing bin collection days
func getMyHousePageFromStroudGov(UPRN int64, client *http.Client) (page string) {
	return
}

// parseMyHousePageForBinDays takes the html document and parses it for the bin collection days. Returns map containing bin types as keys and dates as values, or nil if it could not parse the data
func parseMyHousePageForBinDays(page string) (binDays map[string]time.Time) {
	return
}

//HandleGetBinDayInfoIntent function responsible for the GetBinDayInfoIntent. Takes the request struct, calls relevant functions for calculating the bin day and returns the vale in the Response struct
func HandleGetBinDayInfoIntent(request Request) (resp Response) {
	// Confirm user is in Stroud District Council
	postcodes := strings.Split(os.Getenv("SDC_POSTCODES"), " ")
	deviceID := request.Context.System.Device.DeviceID
	accessToken := request.Context.System.APIAccessToken
	apiEndpoint := request.Context.System.APIEndpoint
	client := &http.Client{}
	postcode, err := getUserPostcode(deviceID, accessToken, apiEndpoint, client)
	if err != nil {
		return AskForPermissionResponse("To retrieve your bin collection day I require your postcode.", []string{"read::alexa:device:all:address:country_and_postal_code"})
	}
	fmt.Println(postcode)

	isInSDC := checkIfPostcodeIsInSDC(postcode, postcodes)
	if !isInSDC {
		return NewSimpleResponse("Cannot fulfill", fmt.Sprintf("I'm sorry, the postcode %s does not belong in Stroud District Council so I cannot look up your bin timetable. Please look for a similar skill in the skill store that is relevant to your area.", postcode))
	}

	sess := session.New()
	svc := dynamodb.New(sess)
	UPRN := getUPRNFromDynamoDB(postcode, svc)
	if UPRN == -1 {
		// UPRN is not in dynamoDB so lookup with API
		UPRN = lookupUPRNForPostcodeViaAPI(postcode, client)
		if UPRN == -1 {
			fmt.Println("Could not get a URPN from the provided postcode")
			return NewSimpleResponse("Cannot fulfill", fmt.Sprintln("I'm sorry, something went wrong getting your property details from the postcode I have recorded against your Amazon device."))
		}

		success := addPostcodeAndUPRNToDynamoDB(postcode, UPRN)
		if !success {
			return NewSimpleResponse("Cannot fulfill", fmt.Sprintln("I'm sorry, something went wrong. Please file a bug to the developer."))
		}

	}
	page := getMyHousePageFromStroudGov(UPRN, client)
	binDays := parseMyHousePageForBinDays(page)

	// formulate map of binDays into an Alexa response
	fmt.Println(binDays)
	return
}

func getUserPostcode(deviceID string, accessToken string, apiEndpoint string, client *http.Client) (result string, err error) {
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	fmt.Println(string(body))
	var add Address
	err = json.Unmarshal(body, &add)
	if err != nil {
		fmt.Println(err)
		return
	}
	result = add.PostalCode
	return
}
func handler(request Request) (Response, error) {
	var response Response

	switch request.Body.Intent.Name {
	case "GetBinDayInfoIntent":
		response = HandleGetBinDayInfoIntent(request)
		break
	default:
		fmt.Println("Other request placeholder")
		break
	}

	fmt.Printf("Starting lambda\n")
	return response, nil
}

func main() {
	lambda.Start(handler)
}

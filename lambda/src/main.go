package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	tableName            = os.Getenv("TABLE_NAME")
	idealPostcodesURL    = os.Getenv("IDEAL_POSTCODES_URL")
	idealPostcodesAPIKey = os.Getenv("IDEAL_POSTCODES_API_KEY")
	stroudGovMyHouseURL  = "https://www.stroud.gov.uk/my-house"
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
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Postcode": {
				S: aws.String(postcode),
			},
		},
	})

	if result.Item == nil {
		return
	}
	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		fmt.Printf("Issue unmarshalling table data, %v\n", err)
	}
	UPRN = item.UPRN
	return
}

// addPostcodeAndUPRNToDynamoDB adds the postcode and corresponding URPN to dynamoDB for future look up. Takes the postcode UPRN and returns true if successful, false if not
func addPostcodeAndUPRNToDynamoDB(postcode string, UPRN int64, svc dynamodbiface.DynamoDBAPI) (result bool) {
	result = false
	entry := Item{
		Postcode: postcode,
		UPRN:     UPRN,
	}

	av, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		fmt.Printf("Error Marshalling Item struct %s\n", err)
		return
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Printf("Error adding item to database: %s\n", err)
		return
	}
	result = true
	return
}

// lookupUPRNForPostcodeViaAPI takes the postcode and calls external API to lookup UPRN. Returns UPRN if successful, -1 if not
func lookupUPRNForPostcodeViaAPI(url string, client *http.Client) (UPRN int64) {
	UPRN = -1
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating http request %v\n", err)
		return
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("Error making http request %v\n", err)
		return
	}

	type IdealResult struct {
		Result []struct {
			UPRN string `json:"uprn"`
		} `json:"result"`
	}

	var ir IdealResult

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body %v\n", err)
		return
	}
	err = json.Unmarshal(body, &ir)
	if err != nil || len(ir.Result) == 0 {
		fmt.Printf("Error unmarshalling JSON %v\n", err)
		return
	}
	UPRN, err = strconv.ParseInt(ir.Result[0].UPRN, 10, 64)
	if err != nil {
		fmt.Printf("Error parsing UPRN %v\n", err)
		return
	}
	if UPRN == 0 {
		UPRN = -1
	}
	return
}

// getMyHousePageFromStroudGov takes the URPN and an http Client and makes an http request to stroud.gov.uk website. Returns the html page containing bin collection days
func getMyHousePageFromStroudGov(UPRN int64, client *http.Client, url string) (page string) {
	page = ""
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Issue creating request %v\n", err)
		return
	}

	cookie := http.Cookie{
		Name:  "myHouse",
		Value: fmt.Sprintf("search=&uprn=%d&address=", UPRN),
	}
	req.AddCookie(&cookie)

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("Error making http request %v\n", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading body %s\n", err)
		return
	}

	return string(body)
}

func getDate(n *html.Node, gardenWaste bool) string {
	var m *html.Node
	if gardenWaste {
		m = n.NextSibling.NextSibling.FirstChild.NextSibling.FirstChild
	} else {
		m = n.NextSibling.NextSibling.FirstChild.FirstChild
	}
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, m)
	return buf.String()
}

// parseMyHousePageForBinDays takes the html document and parses it for the bin collection days. Returns map containing bin types as keys and dates as values, or nil if it could not parse the data
func parseMyHousePageForBinDays(page string) (map[string]string) {
	var binDates = make(map[string]string)
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		fmt.Printf("Error parsing html document %v\n", err)
		return nil
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, a := range n.Attr {
				if a.Key == "alt" {
					switch a.Val {
					case "wheelie-binpng":
						binDates["wheelie bin"] = getDate(n, false)
						break
					case "recycling":
						binDates["recycling"] = getDate(n, false)
						break
					case "food":
						binDates["food"] = getDate(n, false)
						break
					case "fallen-treepng":
						binDates["garden waste"] = getDate(n, true)
						break
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	if binDates["wheelie bin"] == "" || binDates["recycling"] == "" || binDates["food"] == "" | binDates["garden waste" == "" {
		return nil
	}
	return binDates
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
		url := fmt.Sprintf(idealPostcodesURL, postcode, idealPostcodesAPIKey)
		UPRN = lookupUPRNForPostcodeViaAPI(url, client)
		if UPRN == -1 {
			fmt.Println("Could not get a URPN from the provided postcode")
			return NewSimpleResponse("Cannot fulfill", fmt.Sprintln("I'm sorry, something went wrong getting your property details from the postcode I have recorded against your Amazon device."))
		}

		success := addPostcodeAndUPRNToDynamoDB(postcode, UPRN, svc)
		if !success {
			return NewSimpleResponse("Cannot fulfill", fmt.Sprintln("I'm sorry, something went wrong. Please file a bug to the developer."))
		}

	}
	page := getMyHousePageFromStroudGov(UPRN, client, stroudGovMyHouseURL)
	binDays := parseMyHousePageForBinDays(page)
    if binDays == nil {
		return NewSimpleResponse("Cannot fulfill", fmt.Sprintln("I'm sorry, I was not able to get your bin collection data from Stroud District Council's website. Please file a bug to the developer."))
	}
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

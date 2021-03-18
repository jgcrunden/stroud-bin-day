package main

import (
	//	"net/http"
	"errors"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckIfPostcodeIsInSDCWhereItIs(t *testing.T) {
	postcodes := []string{"SW1B", "SW1C", "SW1A", "SW1D"}
	result := checkIfPostcodeIsInSDC("SW1A 2AA", postcodes)
	if result != true {
		t.Error("Expected true, got:", result)
	}
}
func TestCheckIfPostcodeIsInSDCWhereItIsNot(t *testing.T) {
	postcodes := []string{"SW1B", "SW1C", "SW1", "SW1D"}
	result := checkIfPostcodeIsInSDC("SW1A 2AA", postcodes)
	if result != false {
		t.Error("Expected false, got:", result)
	}
}

type mockDynamoDBClientSuccess struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientSuccess) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	response := dynamodb.GetItemOutput{}
	cc := dynamodb.ConsumedCapacity{}
	item := map[string]*dynamodb.AttributeValue{}
	av1 := dynamodb.AttributeValue{}
	av1.SetS("ABC 1234")
	av2 := dynamodb.AttributeValue{}
	av2.SetN("123456789")
	item["Postcode"] = &av1
	item["UPRN"] = &av2
	response.SetConsumedCapacity(&cc)
	response.SetItem(item)
	return &response, nil
}

func TestGetUPRNFromDynamoDBWherePostcodeExists(t *testing.T) {
	mockSvc := mockDynamoDBClientSuccess{}
	result := getUPRNFromDynamoDB("ABC 1234", &mockSvc)
	if result == -1 {
		t.Error("Expected a positive number, got", result)
	}

	var expected int64 = 123456789
	if result != expected {
		t.Error("Expected", expected, ",got ", result)
	}
}

type mockDynamoDBClientFail struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientFail) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	response := dynamodb.GetItemOutput{}
	cc := dynamodb.ConsumedCapacity{}
	response.SetConsumedCapacity(&cc)
	response.SetItem(nil)
	return &response, nil
}

func TestGetUPRNFromDynamoDBWherePostcodeDoesNotExists(t *testing.T) {
	mockSvc := mockDynamoDBClientFail{}
	result := getUPRNFromDynamoDB("ABC 1234", &mockSvc)
	if result != -1 {
		t.Error("Expected a negative number to indicate postcode does not exist, got", result)
	}
}

func (m *mockDynamoDBClientSuccess) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	response := dynamodb.PutItemOutput{}

	return &response, nil
}
func TestAddPostcodeAndUPRNToDynamoDBWhenSuccessful(t *testing.T) {
	mockSvc := mockDynamoDBClientSuccess{}
	result := addPostcodeAndUPRNToDynamoDB("SW1A 2AA", 123456789, &mockSvc)
	if !result {
		t.Error("Expected true to indicate adding postcode and UPRN was successful, got", result)
	}
}

func (m *mockDynamoDBClientFail) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	err := errors.New("ResourceNotFoundException")
	return nil, err
}
func TestAddPostcodeAndUPRNToDynamoDBWhenUnsuccessful(t *testing.T) {
	mockSvc := mockDynamoDBClientFail{}
	result := addPostcodeAndUPRNToDynamoDB("SW1A 2AA", 123456789, &mockSvc)
	if result {
		t.Error("Expected false to indicate adding postcode and UPRN was unsuccessful, got", result)
	}
}

func TestLookupUPRNForPostcodeViaAPIWhereSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{ "result": [ { "postcode": "ID1 1QD", "postcode_inward": "1QD", "postcode_outward": "ID1", "post_town": "LONDON", "dependant_locality": "", "double_dependant_locality": "", "thoroughfare": "Barons Court Road", "dependant_thoroughfare": "", "building_number": "2", "building_name": "", "sub_building_name": "", "po_box": "", "department_name": "", "organisation_name": "", "udprn": 25962203, "umprn": "", "postcode_type": "S", "su_organisation_indicator": "", "delivery_point_suffix": "1G", "line_1": "2 Barons Court Road", "line_2": "", "line_3": "", "premise": "2", "country": "England", "county": "Greater London", "administrative_county": "", "postal_county": "", "traditional_county": "Greater London", "district": "Hammersmith and Fulham", "ward": "North End", "longitude": -0.208644362766368, "latitude": 51.4899488390558, "eastings": 524466, "northings": 178299, "uprn": "2" } ] }`))
	}))
	defer server.Close()

	client := server.Client()
	result := lookupUPRNForPostcodeViaAPI(server.URL, client)
	if result == -1 {
		t.Error("Expected a UPRN, got", result)
	}
}

func TestLookupUPRNForPostcodeViaAPIWhereUnsuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{ "result": [ ] }`))
	}))
	defer server.Close()

	client := server.Client()
	result := lookupUPRNForPostcodeViaAPI(server.URL, client)
	if result != -1 {
		t.Error("Expected a negative UPRN the indicate failed call to API, got", result)
	}
}

var htmlFromStroudGov string = `<html lang="en">
        <head>
        </head>
        <body>
                <div class="panel panel-rubbish">
                        <div class="panel-heading">
                                <h3 class="panel-title">
                                        <span class="fa fa-fw fa-trash-o" style="color: #fff;"></span>Bins, rubbish & recycling
                                </h3>
                        </div>
                        <ul class="list-group" style="padding-left: 0; padding-bottom: 0;">
                                <li class="list-group-item">
                                        <img src="//www.stroud.gov.uk/media/1274/wheelie-bin.png" class="imglandingicon" alt="wheelie-binpng" />
                                        Next rubbish collection date
                                        <p><strong>Tuesday 16 March 2021</strong></p>
                                </li>
                                <li class="list-group-item">
                                        <img src="//www.stroud.gov.uk/media/1266/skip.png" class="imglandingicon" alt="recycling" />
                                        Next recycling collection date
                                        <p><strong>Tuesday 23 March 2021</strong></p>
                                </li>
                                <li class="list-group-item">
                                        <img src="//www.stroud.gov.uk/media/1182/eating.png" class="imglandingicon" alt="food" />
                                        Food waste collection
                                        <p><strong>Every Tuesday</strong></p>
                                </li>
                                <li class="list-group-item">
                                        <img src="//www.stroud.gov.uk/media/1150/fallen-tree.png" class="imglandingicon" alt="fallen-treepng" />
                                        Garden waste collection
                                        <p> <strong>Tuesday 16 March 2021</strong><br />
                                                Find out more about the <a href="/environment/bins-rubbish-and-recycling/garden-waste-collection-service">
                                                        garden waste collection service
                                                </a>
                                        </p>
                                </li>
                                <li class="list-group-item">
                                        <img src="//www.stroud.gov.uk/media/1149/events.png" class="imglandingicon" alt="calendar" />
                                                Download your collection days calendar
                                        <p><a href="https://www.stroud.gov.uk/info/cals2021/Wk2_E2__2021_Tuesday.pdf" target="_blank">Colour calendar</a><br />
                                                <a href="https://www.stroud.gov.uk/info/cals2021/Wk2_E2__2021_Tuesday_BW.pdf" target="_blank">Accessible/printable calendar</a></p>
                                </li>
                                <li class="list-group-item">
                                        <p>
                                                Please ensure that your rubbish & recycling is out by 6am on your collection day.<br />
                                        <a href="/environment/bins-rubbish-and-recycling">More about bins, rubbish & recycling</a>
                                        </p>
                                </li>
                        </ul>
                </div>
        </body>
</html>`

func TestGetMyHousePageFromStroudGovWhereSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(htmlFromStroudGov))
	}))
	defer server.Close()
	client := server.Client()
	result := getMyHousePageFromStroudGov(123456789, client, server.URL)
	if result == "" {
		t.Error("Expected HTML document returned, got an empty string")
	}
}

func TestGetMyHousePageFromStroudGovWhereUnsuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(""))
	}))
	defer server.Close()
	client := server.Client()
	result := getMyHousePageFromStroudGov(123456789, client, server.URL)
	if result != "" {
		t.Error("Expected HTML document returned, got an empty string")
	}
}

func TestParseMyHousePageForBinDaysWhereSuccessful(t *testing.T) {
	result := parseMyHousePageForBinDays(htmlFromStroudGov)
	if result == nil {
		t.Error("Expected a map containing bin types mapped to dates, got", result)
	}
}

func TestParseMyHousePageForBinDaysWhereUnsuccessful(t *testing.T) {
	result := parseMyHousePageForBinDays("")
	if result != nil {
		t.Error("Expected an empty map to indicate unable to parse HTML document", result)
	}
}

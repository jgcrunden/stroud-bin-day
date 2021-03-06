package main

import (
	"testing"
	"net/http"
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
func TestGetUPRNFromDynamoDBWherePostcodeExists(t *testing.T) {
	result := getUPRNFromDynamoDB("ABC 1234")
	if result == -1 {
		t.Error("Expected a positive number, got", result)
	}
}
func TestGetUPRNFromDynamoDBWherePostcodeDoesNotExists(t *testing.T) {
	result := getUPRNFromDynamoDB("ABC 1234")
	if result != -1 {
		t.Error("Expected a negative number to indicate postcode does not exist, got", result)
	}
}
func TestAddPostcodeAndUPRNToDynamoDBWhenSuccessful(t *testing.T) {
	result := addPostcodeAndUPRNToDynamoDB("SW1A 2AA", 123456789)
	if !result {
		t.Error("Expected true to indicate adding postcode and UPRN was successful, got", result)
	}
}
func TestAddPostcodeAndUPRNToDynamoDBWhenUnsuccessful(t *testing.T) {
	result := addPostcodeAndUPRNToDynamoDB("SW1A 2AA", 123456789)
	if result {
		t.Error("Expected false to indicate adding postcode and UPRN was unsuccessful, got", result)
	}
}

func TestLookupUPRNForPostcodeViaAPIWhereSuccessful(t *testing.T) {
	client := &http.Client{}
	result := lookupUPRNForPostcodeViaAPI("SW1A 2AA", client)
	if result == -1 {
		t.Error("Expected a UPRN, got", result)
	}
}

func TestLookupUPRNForPostcodeViaAPIWhereUnsuccessful(t *testing.T) {
	client := &http.Client{}
	result := lookupUPRNForPostcodeViaAPI("SW1A 2AA", client)
	if result != -1 {
		t.Error("Expected -1 to indicate a failed API call", result)
	}
}
func TestGetMyHousePageFromStroudGovWhereSuccessful(t *testing.T) {
	client := &http.Client{}
	result := getMyHousePageFromStroudGov(123456789, client)
	if result == "" {
		t.Error("Expected HTML document returned, got an empty string")
	}
}

func TestGetMyHousePageFromStroudGovWhereUnsuccessful(t *testing.T) {
	client := &http.Client{}
	result := getMyHousePageFromStroudGov(123456789, client)
	if result != "" {
		t.Error("Expected HTML document returned, got an empty string")
	}
}

func TestParseMyHousePageForBinDaysWhereSuccessful(t *testing.T) {
	result := parseMyHousePageForBinDays("PLACEHOLDER FOR HTML document")
	if result == nil {
		t.Error("Expected a map containing bin types mapped to dates, got", result)
	}
}

func TestParseMyHousePageForBinDaysWhereUnsuccessful(t *testing.T) {
	result := parseMyHousePageForBinDays("PLACEHOLDER FOR HTML document")
	if result != nil {
		t.Error("Expected an empty map to indicate unable to parse HTML document", result)
	}
}

package main

import "testing"

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

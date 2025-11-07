package mocks

import "testing"

func TestGetMockCBU_ValidCBU1234(t *testing.T) {
	t.Skip("DEPRECATED: Hardcoded mock data disabled - use database-backed DataStore interface instead")

	cbuID := "CBU-1234"
	cbu, err := GetMockCBU(cbuID)

	// Test now expects deprecation error
	if err == nil {
		t.Fatal("Expected deprecation error, got nil")
	}

	if cbu != nil {
		t.Fatal("Expected nil CBU due to deprecation, got non-nil")
	}

	expectedErrMsg := "DEPRECATED: hardcoded mock data disabled - use database via DataStore interface for CBU_ID: CBU-1234"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected deprecation message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestGetMockCBU_ValidCBU5678(t *testing.T) {
	t.Skip("DEPRECATED: Hardcoded mock data disabled - use database-backed DataStore interface instead")

	cbuID := "CBU-5678"
	cbu, err := GetMockCBU(cbuID)

	// Test now expects deprecation error
	if err == nil {
		t.Fatal("Expected deprecation error, got nil")
	}

	if cbu != nil {
		t.Fatal("Expected nil CBU due to deprecation, got non-nil")
	}

	expectedErrMsg := "DEPRECATED: hardcoded mock data disabled - use database via DataStore interface for CBU_ID: CBU-5678"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected deprecation message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestGetMockCBU_InvalidCBU(t *testing.T) {
	cbuID := "CBU-INVALID"
	cbu, err := GetMockCBU(cbuID)

	if err == nil {
		t.Error("Expected error for invalid CBU ID, got nil")
	}

	if cbu != nil {
		t.Errorf("Expected nil CBU for invalid ID, got: %+v", cbu)
	}

	expectedErrMsg := "DEPRECATED: hardcoded mock data disabled - use database via DataStore interface for CBU_ID: CBU-INVALID"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestGetMockCBU_EmptyCBU(t *testing.T) {
	cbuID := ""
	cbu, err := GetMockCBU(cbuID)

	if err == nil {
		t.Error("Expected error for empty CBU ID, got nil")
	}

	if cbu != nil {
		t.Errorf("Expected nil CBU for empty ID, got: %+v", cbu)
	}
}

func TestCBUStruct_FieldTypes(t *testing.T) {
	cbu := CBU{
		CBUId:         "TEST-123",
		Name:          "Test Fund",
		NaturePurpose: "Test Purpose",
	}

	// Verify all fields are strings
	if cbu.CBUId == "" {
		t.Error("CBUId should be a non-empty string")
	}

	if cbu.Name == "" {
		t.Error("Name should be a non-empty string")
	}

	if cbu.NaturePurpose == "" {
		t.Error("NaturePurpose should be a non-empty string")
	}
}

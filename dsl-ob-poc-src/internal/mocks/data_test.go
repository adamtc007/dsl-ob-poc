package mocks

import "testing"

func TestGetMockCBU_ValidCBU1234(t *testing.T) {
	cbuID := "CBU-1234"
	cbu, err := GetMockCBU(cbuID)

	if err != nil {
		t.Fatalf("Expected no error for valid CBU ID, got: %v", err)
	}

	if cbu == nil {
		t.Fatal("Expected non-nil CBU, got nil")
	}

	if cbu.CBUId != cbuID {
		t.Errorf("Expected CBUId to be '%s', got '%s'", cbuID, cbu.CBUId)
	}

	if cbu.Name != "Aviva Investors Global Fund" {
		t.Errorf("Expected Name to be 'Aviva Investors Global Fund', got '%s'", cbu.Name)
	}

	if cbu.NaturePurpose != "UCITS equity fund domiciled in LU" {
		t.Errorf("Expected NaturePurpose to be 'UCITS equity fund domiciled in LU', got '%s'", cbu.NaturePurpose)
	}
}

func TestGetMockCBU_ValidCBU5678(t *testing.T) {
	cbuID := "CBU-5678"
	cbu, err := GetMockCBU(cbuID)

	if err != nil {
		t.Fatalf("Expected no error for valid CBU ID, got: %v", err)
	}

	if cbu == nil {
		t.Fatal("Expected non-nil CBU, got nil")
	}

	if cbu.CBUId != cbuID {
		t.Errorf("Expected CBUId to be '%s', got '%s'", cbuID, cbu.CBUId)
	}

	if cbu.Name != "Blackrock US Debt Fund" {
		t.Errorf("Expected Name to be 'Blackrock US Debt Fund', got '%s'", cbu.Name)
	}

	if cbu.NaturePurpose != "Corporate debt fund domiciled in IE" {
		t.Errorf("Expected NaturePurpose to be 'Corporate debt fund domiciled in IE', got '%s'", cbu.NaturePurpose)
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

	expectedErrMsg := "no mock data found for CBU_ID: CBU-INVALID"
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

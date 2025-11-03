package mocks

import "fmt"

// CBU represents mock CBU data.
type CBU struct {
	CBUId         string
	Name          string
	NaturePurpose string
}

// GetMockCBU returns mock data for a given CBU ID.
func GetMockCBU(cbuID string) (*CBU, error) {
	if cbuID == "CBU-1234" {
		return &CBU{
			CBUId:         "CBU-1234",
			Name:          "Aviva Investors Global Fund",
			NaturePurpose: "UCITS equity fund domiciled in LU",
		}, nil
	}

	if cbuID == "CBU-5678" {
		return &CBU{
			CBUId:         "CBU-5678",
			Name:          "Blackrock US Debt Fund",
			NaturePurpose: "Corporate debt fund domiciled in IE",
		}, nil
	}

	// Default fallback
	return nil, fmt.Errorf("no mock data found for CBU_ID: %s", cbuID)
}

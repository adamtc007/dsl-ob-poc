package mocks

import "fmt"

// CBU represents mock CBU data.
type CBU struct {
	CBUId         string
	Name          string
	NaturePurpose string
}

// GetMockCBU returns mock data for a given CBU ID.
// COMMENTED OUT: Hardcoded data mocks cause issues with database consistency.
// Use database-backed data instead via DataStore interface.
func GetMockCBU(cbuID string) (*CBU, error) {
	// FIXME: Remove hardcoded mock data - use database instead
	return nil, fmt.Errorf("DEPRECATED: hardcoded mock data disabled - use database via DataStore interface for CBU_ID: %s", cbuID)

	/*
		// DISABLED: Hardcoded mock data
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
	*/
}

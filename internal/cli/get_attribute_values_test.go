package cli

import (
	"testing"
)

// DISABLED: MockStore removed from production code
// This test needs to be rewritten to use real PostgreSQL with test fixtures
// or use a proper database mocking framework like sqlmock
//
// See: internal/store/*_test.go for examples of sqlmock usage
// TODO: Convert to integration test with testcontainers or similar
func TestGetAttributeValues_Integration(t *testing.T) {
	t.Skip("MockStore removed from production code - test needs conversion to real PostgreSQL")

	// OLD TEST CODE REMOVED
	// Previously used datastore.MockStore with JSON files
	// Created temp JSON files and loaded them via MockStore
	// Need to rewrite using:
	// 1. Real PostgreSQL database (testcontainers)
	// 2. OR sqlmock for unit testing
	// 3. OR proper test fixtures in a test database
}

func TestLooksLikeUUID(t *testing.T) {
	validUUID := "123e4567-e89b-12d3-a456-426614174000"
	invalidUUID := "not-a-uuid"

	if !looksLikeUUID(validUUID) {
		t.Errorf("Expected %s to be recognized as UUID", validUUID)
	}

	if looksLikeUUID(invalidUUID) {
		t.Errorf("Expected %s to NOT be recognized as UUID", invalidUUID)
	}
}

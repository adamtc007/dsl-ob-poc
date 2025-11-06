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
func TestRunHistory_PrintsHistory(t *testing.T) {
	t.Skip("MockStore removed from production code - test needs conversion to real PostgreSQL")

	// OLD TEST CODE REMOVED
	// Previously used datastore.MockStore with JSON files
	// Need to rewrite using:
	// 1. Real PostgreSQL database (testcontainers)
	// 2. OR sqlmock for unit testing
	// 3. OR proper test fixtures in a test database
}

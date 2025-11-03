package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"dsl-ob-poc/internal/store"
)

func TestRunHistory_PrintsHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := store.NewStoreFromDB(db)

	cbu := "CBU-1234"
	t1 := time.Now().Add(-2 * time.Minute).Truncate(time.Second)
	t2 := t1.Add(1 * time.Minute)

	rows := sqlmock.NewRows([]string{"version_id", "created_at", "dsl_text"}).
		AddRow("11111111-1111-1111-1111-111111111111", t1, "(dsl v1)").
		AddRow("22222222-2222-2222-2222-222222222222", t2, "(dsl v2)")

	query := regexp.QuoteMeta(`SELECT version_id::text, created_at, dsl_text
         FROM "dsl-ob-poc".dsl_ob
         WHERE cbu_id = $1
         ORDER BY created_at ASC`)
	mock.ExpectQuery(query).WithArgs(cbu).WillReturnRows(rows)

	// Capture stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	err = RunHistory(context.Background(), s, []string{"--cbu", cbu})
	w.Close()
	if err != nil {
		t.Fatalf("RunHistory returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if !regexp.MustCompile(`Found\s+2\s+versions`).MatchString(out) {
		t.Errorf("output did not report 2 versions: %s", out)
	}
	if !regexp.MustCompile(`Version\s+1`).MatchString(out) || !regexp.MustCompile(`Version\s+2`).MatchString(out) {
		t.Errorf("output missing version headers: %s", out)
	}

	if mockErr := mock.ExpectationsWereMet(); mockErr != nil {
		t.Fatalf("unmet sqlmock expectations: %v", mockErr)
	}
}

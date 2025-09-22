package internal

import (
	"bytes"
	"encoding/csv"
	"io"
	"testing"
	"time"
)

// mockWriteCloser wraps a buffer to be used as an io.WriteCloser.
type mockWriteCloser struct {
	bytes.Buffer
}

func (mwc *mockWriteCloser) Close() error {
	// No-op
	return nil
}

func TestLog_Record(t *testing.T) {
	var l Log
	score := IncorrectDifficult // Was ScoreAgain

	before := time.Now()
	l.Record(score)
	after := time.Now()

	if l.Score != score {
		t.Errorf("log.Score = %v, want %v", l.Score, score)
	}

	if l.ReviewedAt.IsZero() {
		t.Error("log.ReviewedAt was not set")
	}

	if l.ReviewedAt.Before(before) || l.ReviewedAt.After(after) {
		t.Errorf("log.ReviewedAt (%v) is not between %v and %v", l.ReviewedAt, before, after)
	}
}

func TestWriteLog(t *testing.T) {
	logBuffer := &mockWriteCloser{}

	deck := &Deck{
		// Correctly initialize LogWriter function
		LogWriter: func() (io.WriteCloser, error) {
			return logBuffer, nil
		},
		// Use []Card instead of []*Card
		Cards: []Card{
			{ // Card 1: Reviewed
				// Use &Meta{} for pointer
				Meta: &Meta{Hash: 123},
				Log: Log{
					ReviewedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Score:      CorrectEasy, // Was ScoreGood
				},
			},
			{ // Card 2: Not reviewed
				Meta: &Meta{Hash: 456},
				Log:  Log{}, // ReviewedAt is zero
			},
			{ // Card 3: Reviewed
				Meta: &Meta{Hash: 789},
				Log: Log{
					ReviewedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
					Score:      CorrectDifficult, // Was ScoreHard
				},
			},
		},
	}

	err := WriteLog(deck)
	if err != nil {
		t.Fatalf("WriteLog() returned an unexpected error: %v", err)
	}

	reader := csv.NewReader(logBuffer)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV output: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 log records, but got %d", len(records))
	}

	// Correct scores in expected output
	// CorrectEasy is 4, CorrectDifficult is 3
	expectedRecord1 := []string{"123", "1672574400000", "4"}
	if !equalSlices(records[0], expectedRecord1) {
		t.Errorf("Record 1 = %v, want %v", records[0], expectedRecord1)
	}

	expectedRecord2 := []string{"789", "1672660800000", "3"}
	if !equalSlices(records[1], expectedRecord2) {
		t.Errorf("Record 2 = %v, want %v", records[1], expectedRecord2)
	}
}

func TestWriteLog_NoReviewedCards(t *testing.T) {
	logBuffer := &mockWriteCloser{}

	deck := &Deck{
		LogWriter: func() (io.WriteCloser, error) {
			return logBuffer, nil
		},
		Cards: []Card{
			{
				Meta: &Meta{Hash: 123},
				Log:  Log{},
			},
			{
				Meta: &Meta{Hash: 456},
				Log:  Log{},
			},
		},
	}

	err := WriteLog(deck)
	if err != nil {
		t.Fatalf("WriteLog() with no reviewed cards returned an unexpected error: %v", err)
	}

	if logBuffer.String() != "" {
		t.Errorf("Expected empty log output, but got: %q", logBuffer.String())
	}
}

// Helper function to compare string slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

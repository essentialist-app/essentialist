package internal

import (
	"os"
	"testing"
	"time"
)

func TestOpenDeck(t *testing.T) {
	d, err := NewDeckFromFile("samples/testdata/test-1.md")
	if err != nil {
		t.Fatal(err)
	}
	cards := d.SelectBefore(time.Now())
	if len(cards) != 5 {
		t.Errorf("Missing cards: %d", len(cards))
	}
}

func TestMissingAnswer(t *testing.T) {
	d, err := NewDeckFromFile("samples/testdata/test-2.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The file has 5 questions, but question 5 has no answer and should be skipped
	if len(d.Cards) != 4 {
		t.Errorf("Expected 4 cards (skipping the one with no answer), got %d", len(d.Cards))
	}
}

func TestCreateDB(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "deck")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write([]byte(`
## question 1
answer 1
## question 2
answer 2
`))
	if err != nil {
		t.Fatalf("write error: %s", err)
	}
	file.Close()

	d, err := NewDeckFromFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	err = d.SaveDeckMeta()
	if err != nil {
		t.Error(err)
	}
}

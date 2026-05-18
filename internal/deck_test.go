package internal

import (
	"io/ioutil"
	"log"
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
	_, err := NewDeckFromFile("samples/testdata/test-2.md")
	if err == nil {
		t.Error("missing error")
	}
}

func TestCreateDB(t *testing.T) {
	file, err := ioutil.TempFile("samples/testdata", "deck")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
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
	os.Remove(file.Name() + ".db")
}

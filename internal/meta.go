package internal

import (
	"encoding/json"
	"errors"
	"hash/fnv"
	"io"
	"strings"
	"time"
)

// Digest represents the question digest.
type Digest uint64

const (
	defaultEasiness = 2.5
	minimumEasiness = 1.3

	FirstRepetitionDelay  = 6  // was 1
	SecondRepetitionDelay = 36 // was 6
)

// Meta contains information about the succces of a card.
type Meta struct {
	Hash       Digest
	NextTime   time.Time // next time to ask
	Repetition int32     // # of success in a row
	Easiness   float32   // how easy is it
}

// NewMeta initialize a new card
func NewMeta(card Card) *Meta {
	return &Meta{
		Hash:       Hash(card),
		Repetition: 0,
		Easiness:   defaultEasiness,
		NextTime:   time.Now(),
	}
}

// Review updates the card meta data according to the score.
// See https://en.wikipedia.org/wiki/SuperMemo
// FirstRepetitionDelay and SecondRepetitionDelay have been
// modified, originally they were 1 and 6.
func (c *Meta) Review(s Score) {
	if s >= 3 {
		switch c.Repetition {
		case 0:
			c.NextTime = time.Now().AddDate(0, 0, FirstRepetitionDelay)
		case 1:
			c.NextTime = time.Now().AddDate(0, 0, SecondRepetitionDelay)
		default:
			// 6 days per successful repetition
			sinceLastTime := float64(c.Repetition) * SecondRepetitionDelay
			days := int(sinceLastTime * float64(c.Easiness))
			c.NextTime = time.Now().AddDate(0, 0, days)
		}

		Q := 5.0 - float32(s)
		c.Easiness = c.Easiness + 0.1 - Q*0.08 - Q*Q*0.02
		if c.Easiness < minimumEasiness {
			c.Easiness = minimumEasiness
		}
		c.Repetition++
	} else {
		c.Repetition = 0
		c.NextTime = time.Now()
	}
}

func strip(s string) string {
	var result strings.Builder
	s = strings.ToLower(s)
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('0' <= b && b <= '9') {
			result.WriteByte(b)
		}
	}
	return result.String()
}

// Hash returns a hash value to index the question. Computed hash is loosy
// since it ignore non alpha numerical values in order to ignore typos
// correction.
func Hash(card Card) Digest {
	h := fnv.New64()
	h.Write([]byte(strip(card.Question)))
	return Digest(h.Sum64())
}

func readDB(r io.Reader) ([]Meta, error) {
	metas := make([]Meta, 0)
	bytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &metas)
	return metas, nil
}

func writeDB(w io.Writer, metas []Meta) error {
	bytes, err := json.MarshalIndent(metas, "", "    ")
	if err != nil {
		return err
	}
	n, err := w.Write(bytes)
	if n != len(bytes) {
		return errors.New("Failed to write DB")
	}
	if err != nil {
		return err
	}
	return nil
}

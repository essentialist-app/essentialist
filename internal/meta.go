package internal

import (
	"encoding/json"
	"errors"
	"hash/fnv"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/open-spaced-repetition/go-fsrs"
)

// Digest represents the question digest.
type Digest uint64

const (
	defaultEasiness = 2.5
	minimumEasiness = 1.3

	FirstRepetitionDelay  = 6  // was 1
	SecondRepetitionDelay = 36 // was 6
)

type Algorithm string

const (
	AlgoSM2  Algorithm = "sm-2"
	AlgoFSRS Algorithm = "fsrs"
)

var CurrentAlgorithm Algorithm = AlgoFSRS

// Meta contains information about the succces of a card.
type Meta struct {
	Hash       Digest
	NextTime   time.Time // next time to ask
	Repetition int32     // # of success in a row
	Easiness   float32   // how easy is it

	// FSRS fields
	FSRSNextTime      time.Time `json:"FSRSNextTime,omitempty"`
	FSRSStability     float64   `json:"FSRSStability,omitempty"`
	FSRSDifficulty    float64   `json:"FSRSDifficulty,omitempty"`
	FSRSReps          int32     `json:"FSRSReps,omitempty"`
	FSRSLapses        int32     `json:"FSRSLapses,omitempty"`
	FSRSState         int8      `json:"FSRSState,omitempty"`
	FSRSLastReview    time.Time `json:"FSRSLastReview,omitempty"`
	FSRSScheduledDays int32     `json:"FSRSScheduledDays,omitempty"`
	FSRSElapsedDays   int32     `json:"FSRSElapsedDays,omitempty"`
}

func (c *Meta) GetNextTime() time.Time {
	if CurrentAlgorithm == AlgoFSRS {
		if c.FSRSNextTime.IsZero() {
			return c.NextTime
		}
		return c.FSRSNextTime
	}
	return c.NextTime
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
	if CurrentAlgorithm == AlgoFSRS {
		c.ReviewFSRS(s)
		return
	}
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

func (c *Meta) ReviewFSRS(s Score) {
	var rating fsrs.Rating
	switch s {
	case 0, 1, 2:
		rating = fsrs.Again
	case 3:
		rating = fsrs.Hard
	case 4:
		rating = fsrs.Good
	case 5:
		rating = fsrs.Easy
	default:
		rating = fsrs.Good
	}

	card := fsrs.Card{
		Due:           c.FSRSNextTime,
		Stability:     c.FSRSStability,
		Difficulty:    c.FSRSDifficulty,
		ElapsedDays:   uint64(c.FSRSElapsedDays),
		ScheduledDays: uint64(c.FSRSScheduledDays),
		Reps:          uint64(c.FSRSReps),
		Lapses:        uint64(c.FSRSLapses),
		State:         fsrs.State(c.FSRSState),
		LastReview:    c.FSRSLastReview,
	}

	now := time.Now()
	if c.FSRSLastReview.IsZero() {
		card.Due = now
		card.State = fsrs.New
	} else {
		elapsed := now.Sub(c.FSRSLastReview).Hours() / 24.0
		if elapsed < 0 {
			elapsed = 0
		}
		card.ElapsedDays = uint64(elapsed)
	}

	params := fsrs.DefaultParam()
	schedulingChoices := params.Repeat(card, now)
	choice := schedulingChoices[rating]

	c.FSRSNextTime = choice.Card.Due
	c.FSRSStability = choice.Card.Stability
	c.FSRSDifficulty = choice.Card.Difficulty
	c.FSRSReps = int32(choice.Card.Reps)
	c.FSRSLapses = int32(choice.Card.Lapses)
	c.FSRSState = int8(choice.Card.State)
	c.FSRSLastReview = choice.Card.LastReview
	c.FSRSScheduledDays = int32(choice.Card.ScheduledDays)
	c.FSRSElapsedDays = int32(choice.Card.ElapsedDays)
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
	bytes, err := ioutil.ReadAll(r)
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

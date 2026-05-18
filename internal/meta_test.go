package internal

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs"
)

var metaInput = []Meta{
	{
		Hash:       0,
		NextTime:   time.Unix(0, 0),
		Repetition: 0,
		Easiness:   0.0,
	},
	{
		Hash:       1,
		NextTime:   time.Unix(2, 0),
		Repetition: 1,
		Easiness:   1.0,
	},
	{
		Hash:       1,
		NextTime:   time.Unix(3, 0),
		Repetition: 1,
		Easiness:   2.0,
	},
}

func TestStrip(t *testing.T) {
	input := []string{
		"\n",
		"\t",
		" ",
		"A",
		"A?",
		"This \n is a long test",
	}
	expected := []string{
		"",
		"",
		"",
		"a",
		"a",
		"thisisalongtest",
	}
	for i, in := range input {
		out := strip(in)
		if out != expected[i] {
			t.Errorf("%d: %s instead of %s", i, out, expected[i])
		}
	}
}

func TestHash(t *testing.T) {
	input := []string{
		"Question a",
		"question A",
		"question A???",
	}
	var card Card
	card.Question = "question a"
	expected := Hash(card)
	for _, in := range input {
		card.Question = in
		out := Hash(card)
		if out != expected {
			t.Errorf("%v instead of %v", out, expected)
		}
	}
}

func TestMetaReview(t *testing.T) {
	var card Card
	meta := NewMeta(card)

	if meta.FSRSReps != 0 {
		t.Errorf("Invalid FSRSReps: %d", meta.FSRSReps)
	}

	meta.Review(ScoreAgain) // Again (1)
	if meta.FSRSReps != 1 {
		t.Errorf("Expected FSRSReps to be 1 after review, got %d", meta.FSRSReps)
	}
	if meta.FSRSNextTime.IsZero() {
		t.Errorf("Expected FSRSNextTime to be set")
	}

	meta.Review(ScoreGood) // Good (3)
	if meta.FSRSReps != 2 {
		t.Errorf("Expected FSRSReps to be 2 after review, got %d", meta.FSRSReps)
	}
}

func TestWriteRead(t *testing.T) {
	var buf bytes.Buffer
	err := writeDB(&buf, metaInput)
	if err != nil {
		t.Fatal(err)
	}
	output, err := readDB(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(metaInput) != len(output) {
		t.Errorf("len %d, expected: %d", len(output), len(metaInput))
	}
	for i := range output {
		if metaInput[i].Hash != output[i].Hash {
			t.Errorf("%d, Hash: %d / %d", i,
				output[i].Hash, metaInput[i].Hash)
		}
	}
}

func TestFSRSMetaReview(t *testing.T) {
	var card Card
	meta := NewMeta(card)

	// In initial state, FSRS fields are zero values
	if !meta.FSRSNextTime.IsZero() {
		t.Errorf("FSRSNextTime should be zero: %v", meta.FSRSNextTime)
	}
	if meta.FSRSReps != 0 {
		t.Errorf("FSRSReps should be 0: %d", meta.FSRSReps)
	}

	// Perform a review under FSRS
	meta.Review(ScoreEasy) // Easy rating (4)

	if meta.FSRSReps != 1 {
		t.Errorf("FSRSReps should be 1: %d", meta.FSRSReps)
	}
	if meta.FSRSNextTime.IsZero() {
		t.Errorf("FSRSNextTime should not be zero")
	}
	if meta.FSRSDifficulty == 0 {
		t.Errorf("FSRSDifficulty should not be zero")
	}
	if meta.FSRSStability == 0 {
		t.Errorf("FSRSStability should not be zero")
	}
}

func TestGetNextTime(t *testing.T) {
	var card Card
	meta := NewMeta(card)
	meta.NextTime = time.Now().Add(24 * time.Hour)
	meta.FSRSNextTime = time.Now().Add(48 * time.Hour)

	// Check FSRS next time is returned
	if !meta.GetNextTime().Equal(meta.FSRSNextTime) {
		t.Errorf("Expected FSRS next time %v, got %v", meta.FSRSNextTime, meta.GetNextTime())
	}

	// Check FSRS Zero value behavior falls back to legacy NextTime
	meta.FSRSNextTime = time.Time{}
	if !meta.GetNextTime().Equal(meta.NextTime) {
		t.Errorf("Expected GetNextTime to fall back to NextTime %v, got %v", meta.NextTime, meta.GetNextTime())
	}
}

func TestFSRSScenariosTableDriven(t *testing.T) {
	p := fsrs.DefaultParam()
	w := p.W

	scenarios := []struct {
		name    string
		card    fsrs.Card
		elapsed int
		rating  fsrs.Rating
		verify  func(t *testing.T, name string, result fsrs.Card)
	}{
		{
			name: "Brand New Card - Again",
			card: fsrs.Card{
				State: fsrs.New,
			},
			elapsed: 0,
			rating:  fsrs.Again,
			verify: func(t *testing.T, name string, result fsrs.Card) {
				if math.Abs(result.Stability-w[0]) > 1e-6 {
					t.Errorf("[%s] stability = %f, expected = %f", name, result.Stability, w[0])
				}
				expectedD := w[4] + 2*w[5]
				if math.Abs(result.Difficulty-expectedD) > 1e-6 {
					t.Errorf("[%s] difficulty = %f, expected = %f", name, result.Difficulty, expectedD)
				}
			},
		},
		{
			name: "Brand New Card - Good",
			card: fsrs.Card{
				State: fsrs.New,
			},
			elapsed: 0,
			rating:  fsrs.Good,
			verify: func(t *testing.T, name string, result fsrs.Card) {
				if math.Abs(result.Stability-w[2]) > 1e-6 {
					t.Errorf("[%s] stability = %f, expected = %f", name, result.Stability, w[2])
				}
				expectedD := w[4]
				if math.Abs(result.Difficulty-expectedD) > 1e-6 {
					t.Errorf("[%s] difficulty = %f, expected = %f", name, result.Difficulty, expectedD)
				}
			},
		},
		{
			name: "Brand New Card - Easy",
			card: fsrs.Card{
				State: fsrs.New,
			},
			elapsed: 0,
			rating:  fsrs.Easy,
			verify: func(t *testing.T, name string, result fsrs.Card) {
				if math.Abs(result.Stability-w[3]) > 1e-6 {
					t.Errorf("[%s] stability = %f, expected = %f", name, result.Stability, w[3])
				}
				expectedD := w[4] - w[5]
				if math.Abs(result.Difficulty-expectedD) > 1e-6 {
					t.Errorf("[%s] difficulty = %f, expected = %f", name, result.Difficulty, expectedD)
				}
				if result.State != fsrs.Review {
					t.Errorf("[%s] expected state Review (2), got %d", name, result.State)
				}
			},
		},
		{
			name: "Lapse in Review State",
			card: fsrs.Card{
				Stability:  5.0,
				Difficulty: 5.0,
				State:      fsrs.Review,
				Reps:       1,
			},
			elapsed: 5,
			rating:  fsrs.Again,
			verify: func(t *testing.T, name string, result fsrs.Card) {
				if result.Stability >= 5.0 {
					t.Errorf("[%s] expected stability < 5.0 after lapse, got %f", name, result.Stability)
				}
			},
		},
		{
			name: "Clamping of Difficulty - Lower Bound",
			card: fsrs.Card{
				Stability:  5.0,
				Difficulty: 1.1,
				State:      fsrs.Review,
				Reps:       1,
			},
			elapsed: 5,
			rating:  fsrs.Easy,
			verify: func(t *testing.T, name string, result fsrs.Card) {
				if result.Difficulty < 1.0 || result.Difficulty > 10.0 {
					t.Errorf("[%s] expected difficulty clamped between 1 and 10, got %f", name, result.Difficulty)
				}
			},
		},
		{
			name: "Clamping of Difficulty - Upper Bound",
			card: fsrs.Card{
				Stability:  5.0,
				Difficulty: 9.9,
				State:      fsrs.Review,
				Reps:       1,
			},
			elapsed: 5,
			rating:  fsrs.Again,
			verify: func(t *testing.T, name string, result fsrs.Card) {
				if result.Difficulty < 1.0 || result.Difficulty > 10.0 {
					t.Errorf("[%s] expected difficulty clamped between 1 and 10, got %f", name, result.Difficulty)
				}
			},
		},
		{
			name: "Same-Day Review Interval",
			card: fsrs.Card{
				Stability:  5.0,
				Difficulty: 5.0,
				State:      fsrs.Review,
				Reps:       2,
			},
			elapsed: 0,
			rating:  fsrs.Good,
			verify: func(t *testing.T, name string, result fsrs.Card) {
				if result.ScheduledDays == 0 {
					t.Errorf("[%s] expected positive scheduled days, got 0", name)
				}
			},
		},
		{
			name: "Fundamental Mathematical Identity (I == round(S))",
			card: fsrs.Card{
				Stability:  10.0,
				Difficulty: 5.0,
				State:      fsrs.Review,
				Reps:       1,
			},
			elapsed: 10,
			rating:  fsrs.Good,
			verify: func(t *testing.T, name string, result fsrs.Card) {
				expectedI := uint64(math.Round(result.Stability))
				if expectedI < 1 {
					expectedI = 1
				}
				if result.ScheduledDays != expectedI {
					t.Errorf("[%s] expected interval %d, got %d", name, expectedI, result.ScheduledDays)
				}
			},
		},
		{
			name: "Mean Reversion of Difficulty",
			card: fsrs.Card{
				Stability:  10.0,
				Difficulty: 9.0,
				State:      fsrs.Review,
				Reps:       1,
			},
			elapsed: 10,
			rating:  fsrs.Good,
			verify: func(t *testing.T, name string, result fsrs.Card) {
				// Simulate repeated reviews
				currCard := result
				for i := 0; i < 14; i++ {
					currCard.LastReview = time.Now().AddDate(0, 0, -int(currCard.ScheduledDays))
					choices := p.Repeat(currCard, time.Now())
					currCard = choices[fsrs.Good].Card
				}
				if currCard.Difficulty >= 9.0 {
					t.Errorf("[%s] expected difficulty to decrease below 9.0 due to mean reversion, got %f", name, currCard.Difficulty)
				}
			},
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now()
			if tc.card.State != fsrs.New {
				tc.card.LastReview = now.AddDate(0, 0, -tc.elapsed)
			}
			choices := p.Repeat(tc.card, now)
			tc.verify(t, tc.name, choices[tc.rating].Card)
		})
	}
}

func TestFSRSSpacingEffectComparison(t *testing.T) {
	p := fsrs.DefaultParam()

	// Card A (Early Review: 1 day elapsed)
	cardA := fsrs.Card{
		Stability:  5.0,
		Difficulty: 5.0,
		State:      fsrs.Review,
		Reps:       1,
		LastReview: time.Now().AddDate(0, 0, -1),
	}
	choicesA := p.Repeat(cardA, time.Now())
	stabilityA := choicesA[fsrs.Good].Card.Stability

	// Card B (Late Review: 10 days elapsed)
	cardB := fsrs.Card{
		Stability:  5.0,
		Difficulty: 5.0,
		State:      fsrs.Review,
		Reps:       1,
		LastReview: time.Now().AddDate(0, 0, -10),
	}
	choicesB := p.Repeat(cardB, time.Now())
	stabilityB := choicesB[fsrs.Good].Card.Stability

	if stabilityB <= stabilityA {
		t.Errorf("Spacing effect violated: stability for late review (%f) should be greater than early review (%f)", stabilityB, stabilityA)
	}
}

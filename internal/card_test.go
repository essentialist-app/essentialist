package internal

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
)

func TestSplitDeck(t *testing.T) {
	input := `
## L1 cache reference
0.5 ns

##Branch mispredict
5 ns
## L2 cache reference

7 ns,  14x L1 cache
`
	expected := []Card{
		{
			Question: "L1 cache reference",
			Answer:   "0.5 ns",
		},
		{
			Question: "Branch mispredict",
			Answer:   "5 ns",
		},
		{
			Question: "L2 cache reference",
			Answer:   "7 ns,  14x L1 cache",
		},
	}
	cards, err := readCards(bytes.NewBufferString(input), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != len(expected) {
		t.Fatalf("Wrong length: %d", len(cards))
	}
	for i, card := range cards {
		if card.Question != expected[i].Question {
			t.Errorf("Question: %s, instead of: %s",
				card.Question, expected[i].Question)
		}
		if card.Answer != expected[i].Answer {
			t.Errorf("Answer: %s, instead of: %s",
				card.Answer, expected[i].Answer)
		}
	}
}

func TestSplitCardWithCode(t *testing.T) {
	template := `## Title 1

Text 1

## Show me some code

Some text

%s
## A comment
Some code
%s

## Title 3

Text 3
`
	deck := fmt.Sprintf(template, "```", "```")

	cards, lines, _ := splitCards(deck)
	if len(cards) != 3 {
		t.Errorf("Wrong size: %d", len(cards))
	}
	if len(cards) != len(lines) {
		t.Errorf("Each card should have a line: %d vs %d", len(cards), len(lines))
	}
	if cards[1] != fmt.Sprintf(`## Show me some code

Some text

%s
## A comment
Some code
%s
`, "```", "```") {
		t.Errorf("Wrong card: %s", cards[1])
	}
}

func TestMultipleHeaderLevels(t *testing.T) {
	input := `# H1 Question
This is an H1 answer

## H2 Question
This is an H2 answer

### H3 Question
This is an H3 answer

#### H4 Question
This is an H4 answer
`
	expected := []Card{
		{
			Question: "H1 Question",
			Answer:   "This is an H1 answer",
		},
		{
			Question: "H1 Question ‒> H2 Question",
			Answer:   "This is an H2 answer",
		},
		{
			Question: "H1 Question ‒> H2 Question ‒> H3 Question",
			Answer:   "This is an H3 answer",
		},
		{
			Question: "H1 Question ‒> H2 Question ‒> H3 Question ‒> H4 Question",
			Answer:   "This is an H4 answer",
		},
	}
	cards, err := readCards(bytes.NewBufferString(input), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != len(expected) {
		t.Fatalf("Wrong length: got %d, want %d", len(cards), len(expected))
	}
	for i, card := range cards {
		if card.Question != expected[i].Question {
			t.Errorf("Card %d Question: got %q, want %q",
				i, card.Question, expected[i].Question)
		}
		if card.Answer != expected[i].Answer {
			t.Errorf("Card %d Answer: got %q, want %q",
				i, card.Answer, expected[i].Answer)
		}
	}
}

func TestHierarchicalQuestions(t *testing.T) {
	input := `# Main Topic
Some intro text

## Subtopic 1
Answer for subtopic 1

### Detail 1
Answer for detail 1

### Detail 2
Answer for detail 2

## Subtopic 2
Answer for subtopic 2
`
	expected := []Card{
		{
			Question: "Main Topic",
			Answer:   "Some intro text",
		},
		{
			Question: "Main Topic ‒> Subtopic 1",
			Answer:   "Answer for subtopic 1",
		},
		{
			Question: "Main Topic ‒> Subtopic 1 ‒> Detail 1",
			Answer:   "Answer for detail 1",
		},
		{
			Question: "Main Topic ‒> Subtopic 1 ‒> Detail 2",
			Answer:   "Answer for detail 2",
		},
		{
			Question: "Main Topic ‒> Subtopic 2",
			Answer:   "Answer for subtopic 2",
		},
	}
	cards, err := readCards(bytes.NewBufferString(input), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != len(expected) {
		t.Fatalf("Wrong length: got %d, want %d", len(cards), len(expected))
	}
	for i, card := range cards {
		if card.Question != expected[i].Question {
			t.Errorf("Card %d Question: got %q, want %q",
				i, card.Question, expected[i].Question)
		}
		if card.Answer != expected[i].Answer {
			t.Errorf("Card %d Answer: got %q, want %q",
				i, card.Answer, expected[i].Answer)
		}
	}
}

func TestReadCardsWithImagePath(t *testing.T) {
	input := `
## Card with image in answer

This is an image in the answer: ![alt text](images/test.png)

## Card with image in question ![alt text](./images/test2.png)

This is the answer.

## Card with parent dir image

![alt text](../images/test3.png)
`
	expectedCards := []Card{
		{
			Question: "Card with image in answer",
			Answer:   "This is an image in the answer: ![alt text](/path/to/deck/images/test.png)",
		},
		{
			Question: "Card with image in question ![alt text](/path/to/deck/images/test2.png)",
			Answer:   "This is the answer.",
		},
		{
			Question: "Card with parent dir image",
			Answer:   "![alt text](/path/to/images/test3.png)",
		},
	}

	cards, err := readCards(bytes.NewBufferString(input), "/path/to/deck/deck.md")
	if err != nil {
		t.Fatal(err)
	}

	if len(cards) != len(expectedCards) {
		t.Fatalf("Wrong length: got %d, want %d", len(cards), len(expectedCards))
	}

	for i, card := range cards {
		expectedCard := expectedCards[i]

		actualQuestion := filepath.ToSlash(card.Question)
		expectedQuestion := filepath.ToSlash(expectedCard.Question)
		if actualQuestion != expectedQuestion {
			t.Errorf("Card %d Question: got %q, want %q", i, actualQuestion, expectedQuestion)
		}

		actualAnswer := filepath.ToSlash(card.Answer)
		expectedAnswer := filepath.ToSlash(expectedCard.Answer)
		if actualAnswer != expectedAnswer {
			t.Errorf("Card %d Answer: got %q, want %q", i, actualAnswer, expectedAnswer)
		}
	}
}

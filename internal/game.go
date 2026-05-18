package internal

import (
	"fmt"
	"time"
)

// Game represents a learning session.
type Game struct {
	cards    []Card
	decks    []*Deck
	index    int
	success  int
	total    int
	finished bool
}

const (
	ALL_CARDS       = -1
	CARDS_TO_REVIEW = 0
)

// NewGameFromFiles reads the markdown files to instantiate a Game. cardsNb
// represents the maximum number of cards to use.
func NewGameFromFiles(cardsNb int, files []string) (*Game, error) {
	decks := make([]*Deck, len(files))
	for i, file := range files {
		deck, err := NewDeckFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("Failed to load %s: %v", file, err)
		}
		decks[i] = deck
	}
	return NewGame(cardsNb, decks...), nil
}

// NewGame returns a game given a set of markdown files.
//
// If cardsNb is equal to ALL_CARDS, all cards in the deck are used. If cardsNb
// is CARDS_TO_REVIEW then all the cards that needs to be review will be uesd.
// If cardsNb is a strictly positive number, up to cardsNb from the cards to
// review will be used.
func NewGame(cardsNb int, decks ...*Deck) *Game {
	game := &Game{
		cards: make([]Card, 0),
		decks: decks,
	}
	for i, deck := range decks {
		var cards []Card
		if cardsNb == ALL_CARDS {
			cards = deck.Cards
		} else {
			cards = deck.SelectBefore(time.Now())
		}
		game.cards = append(game.cards, cards...)
		game.success += len(deck.Cards) - len(cards)
		game.total += len(deck.Cards)
		game.decks[i] = deck
	}
	game.cards = ShuffleCards(game.cards)
	if cardsNb > 0 && len(game.cards) > cardsNb {
		game.cards = game.cards[0:cardsNb]
	}
	return game
}

func (g *Game) CurrentCard() *Card {
	if len(g.cards) == 0 {
		return nil
	}
	return &g.cards[g.index]
}

// Question returns the next question to answer. Idempotent.
func (g *Game) Question() string {
	if len(g.cards) == 0 {
		return "No cards"
	}
	return g.cards[g.index].Question
}

// Question returns the next question to answer. Idempotent.
func (g *Game) Answer() string {
	if len(g.cards) == 0 {
		return "No cards"
	}
	return g.cards[g.index].Answer
}

func (g *Game) DeckName() string {
	if len(g.cards) == 0 {
		return "zero"
	}
	return g.cards[g.index].DeckName
}

// Score represents how easily one responded to a question.
type Score int

const (
	// 1: Again (fail)
	ScoreAgain Score = 1
	// 2: Hard
	ScoreHard Score = 2
	// 3: Good
	ScoreGood Score = 3
	// 4: Easy
	ScoreEasy Score = 4
)

func (g *Game) Review(s Score) {
	if g.index < len(g.cards) {
		if s >= ScoreHard {
			g.success++
		}
		g.cards[g.index].Meta.Review(s)
		g.index++
	}
	if g.index == len(g.cards) {
		g.index = 0
		g.finished = true
	}
}

func (g *Game) Previous() {
	if g.index > 0 {
		g.index--
	}
}

func (g *Game) Skip() {
	if g.index < len(g.cards) {
		g.index++
	}
	if g.index == len(g.cards) {
		g.index = 0
		g.finished = true
	}
}

func (g *Game) Progress() (current, total int) {
	return g.index + 1, len(g.cards)
}

func (g *Game) Success() float32 {
	if g.total == 0 {
		return 100
	}
	return (float32(g.success) / float32(g.total)) * 100
}

func (g *Game) IsFinished() bool {
	if len(g.cards) == 0 {
		return true
	}
	return g.finished
}

func (g *Game) Save() {
	for _, d := range g.decks {
		defer d.SaveDeckMeta()
	}
}

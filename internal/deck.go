package internal

import (
	"fmt"
	"io"
	"math/rand"
	"time"
)

// MetaMap associate the index if the cards with its meta data.
type MetaMap map[Digest]*Meta

type Deck struct {
	Cards      []Card
	Name       string
	MetaWriter func() (io.WriteCloser, error)
	LogWriter func() (io.WriteCloser, error)
}

func loadCards(accessor DeckAccessor) ([]Card, error) {
	cardReader, err := accessor.CardsReader()
	if err != nil {
		return nil, err
	}
	defer cardReader.Close()

	cards, err := readCards(cardReader)
	if err != nil {
		return nil, err
	}
	return cards, nil
}

func newMetaCards(accessor DeckAccessor) error {
	metas := make([]Meta, 0)
	metaWriter, err := accessor.MetaWriter()
	if err != nil {
		return fmt.Errorf("cannot create DB: %v", err)
	}
	defer metaWriter.Close()

	if err = writeDB(metaWriter, metas); err != nil {
		return fmt.Errorf("Cannot write DB: %v", err)
	}
	return nil
}

func loadMetaMap(accessor DeckAccessor) (MetaMap, error) {
	var metaMap MetaMap = make(map[Digest]*Meta)

	metaReader, err := accessor.MetaReader()
	if err != nil {
		err = newMetaCards(accessor)
		if err != nil {
			return nil, err
		}
		return metaMap, nil
	}
	defer metaReader.Close()

	metas, err := readDB(metaReader)
	if err != nil {
		err = newMetaCards(accessor)
		if err != nil {
			return nil, err
		}
		return metaMap, nil
	}

	for i := range metas {
		metaMap[metas[i].Hash] = &metas[i]
	}
	return metaMap, nil
}

// NewDeckFromFile reads a Deck from a file.
func NewEmptyDeck(name string) *Deck {
	return &Deck{
		Cards:      []Card{},
		Name:       name,
		MetaWriter: func() (io.WriteCloser, error) { return nil, fmt.Errorf("Dummy MetaWriter for %s", name) },
		LogWriter: func() (io.WriteCloser, error) { return nil, fmt.Errorf("Dummy LogWriter for %s", name) },
	}
}

// NewDeckFromFile reads a Deck from a file.
func NewDeckFromFile(filename string) (*Deck, error) {
	return NewDeck(newFileDeckAccessor(filename))
}

// NewDeck reads a Deck from DeckAccessor
func NewDeck(accessor DeckAccessor) (*Deck, error) {
	cards, err := loadCards(accessor)
	if err != nil {
		return nil, err
	}
	metaMap, err := loadMetaMap(accessor)
	if err != nil {
		return nil, err
	}

	deckName := accessor.DeckName()
	for i := range cards {
		cards[i].DeckName = deckName
		hash := Hash(cards[i])
		meta, ok := metaMap[hash]
		if ok {
			cards[i].Meta = meta
		} else {
			cards[i].Meta = NewMeta(cards[i])
		}
	}

	return &Deck{
		Cards:      cards,
		Name:       accessor.DeckName(),
		MetaWriter: accessor.MetaWriter,
		LogWriter:  accessor.LogWriter,
	}, nil
}

func ShuffleCards(cards []Card) []Card {
	rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
	return cards
}

// SelectBefore returns the cards to be review before a given date.
func (d *Deck) SelectBefore(now time.Time) []Card {
	cards := []Card{}
	for _, card := range d.Cards {
		if card.Meta.NextTime.Before(now) {
			cards = append(cards, card)
		}
	}
	return cards
}

func (d *Deck) Stats() (toReview, total int) {
	toReview = 0
	now := time.Now()
	for _, card := range d.Cards {
		if card.Meta.NextTime.Before(now) {
			toReview++
		}
	}
	return toReview, len(d.Cards)
}

func (d *Deck) SaveDeckMeta() error {
	metas := make([]Meta, len(d.Cards))
	for i := range d.Cards {
		metas[i] = *d.Cards[i].Meta
	}
	metaWriter, err := d.MetaWriter()
	if err != nil {
		return err
	}
	defer metaWriter.Close()
	return writeDB(metaWriter, metas)
}

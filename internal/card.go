package internal

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	errCardEmpty       = errors.New("Empty note")
	errQuestionMissing = errors.New("Missing question")
	errInvalidCard     = errors.New("Invalid card")
)

var splitQuestion = regexp.MustCompile(`(?m)^#+\s*`)

type Card struct {
	Question string
	Answer   string
	DeckName string
	Deck     *Deck
	Meta     *Meta
}

func (c Card) Review(s Score) {
	c.Meta.Review(s)
}

// getHeaderLevel returns the level of a header line (number of # characters)
func getHeaderLevel(line string) int {
	level := 0
	for i := 0; i < len(line) && line[i] == '#'; i++ {
		level++
	}
	return level
}

// splitCards take a mardown string as input and returns a set of cards, their line numbers, and hierarchy paths.
func splitCards(md string) ([]string, []int, []string) {
	cards := make([]string, 0)
	cardsLineNb := make([]int, 0)
	cardsHierarchy := make([]string, 0)
	isCode := false // true when parsing "```"
	card := ""      // current card being parsed
	cardLineNb := 0 // current card line number
	lines := strings.Split(md, "\n")

	// Track the header hierarchy as we parse
	// hierarchy[level] = header text at that level
	hierarchy := make(map[int]string)

	for i, line := range lines {
		if splitQuestion.Match([]byte(line)) && !isCode {
			// 1. add previous card to the deck if any.
			// 2. start the card with the title.
			if card != "" {
				cards = append(cards, card)
				cardsLineNb = append(cardsLineNb, cardLineNb)
				// Build hierarchy path for previous card
				hierarchyPath := buildHierarchyPath(hierarchy)
				cardsHierarchy = append(cardsHierarchy, hierarchyPath)
			}

			// Update hierarchy for the new header
			level := getHeaderLevel(line)
			headerText := strings.TrimSpace(line[level:])

			// Update hierarchy at this level and clear deeper levels
			hierarchy[level] = headerText
			for l := level + 1; l <= 6; l++ {
				delete(hierarchy, l)
			}

			cardLineNb = i
			card = line
		} else {
			if strings.HasPrefix(line, "```") {
				isCode = !isCode
			}
			// If this isn't a title, add it to the card.
			card = fmt.Sprintf("%s\n%s", card, line)
		}
	}
	if card != "" {
		cards = append(cards, card)
		cardsLineNb = append(cardsLineNb, cardLineNb)
		// Build hierarchy path for last card
		hierarchyPath := buildHierarchyPath(hierarchy)
		cardsHierarchy = append(cardsHierarchy, hierarchyPath)
	}
	return cards, cardsLineNb, cardsHierarchy
}

// buildHierarchyPath builds a hierarchy path string from the hierarchy map
func buildHierarchyPath(hierarchy map[int]string) string {
	if len(hierarchy) == 0 {
		return ""
	}

	// Find all levels and sort them
	levels := make([]int, 0, len(hierarchy))
	for level := range hierarchy {
		levels = append(levels, level)
	}

	// Simple bubble sort since we have at most 6 levels
	for i := 0; i < len(levels); i++ {
		for j := i + 1; j < len(levels); j++ {
			if levels[i] > levels[j] {
				levels[i], levels[j] = levels[j], levels[i]
			}
		}
	}

	// Build the path
	path := ""
	for i, level := range levels {
		if i > 0 {
			path += " ‒> "
		}
		path += hierarchy[level]
	}
	return path
}

func parseCards(md string, deckPath string) ([]Card, error) {
	cards := make([]Card, 0)

	sheets, lines, hierarchies := splitCards(md)
	for i, sheet := range sheets {
		card, err := loadCard(sheet, deckPath, hierarchies[i])
		if err == errCardEmpty {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("%w (line %d)", err, lines[i])
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func readCards(r io.Reader, deckPath string) ([]Card, error) {
	dat, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return parseCards(string(dat), deckPath)
}

func trim(s string) string {
	return strings.TrimSpace(strings.Trim(s, "\n"))
}

// loadCard parse a card description
func loadCard(md string, deckPath string, hierarchyPath string) (c Card, err error) {
	md = trim(md)
	if md == "" {
		return c, errCardEmpty
	}
	sheets := strings.SplitN(md, "\n", 2)
	if len(sheets) < 1 {
		return c, errInvalidCard
	}
	if !strings.HasPrefix(sheets[0], "#") {
		return c, errInvalidCard
	}
	// Count and remove the '#' characters from the question.
	headerLen := 0
	for i := 0; i < len(sheets[0]) && sheets[0][i] == '#'; i++ {
		headerLen++
	}
	c.Question = trim(sheets[0][headerLen:])
	if c.Question == "" {
		return c, errInvalidCard
	}

	// Prepend hierarchy path if it exists
	if hierarchyPath != "" {
		c.Question = hierarchyPath
	}
	// If there's no content after the header, skip this card
	if len(sheets) == 2 {
		c.Answer = trim(sheets[1])
	} else {
		c.Answer = ""
	}

	// Skip cards with no answer content
	if c.Answer == "" {
		return c, errCardEmpty
	}

	c.Question, err = rewriteImagePaths(c.Question, deckPath)
	if err != nil {
		return c, err
	}
	c.Answer, err = rewriteImagePaths(c.Answer, deckPath)
	if err != nil {
		return c, err
	}

	c.Meta = NewMeta(c)
	return c, nil
}

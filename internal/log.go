package internal

import (
	"time"
)

// Log records what what time a card was reviewed and what score it had
type Log struct {
	ReviewedAt  time.Time // time at which it was reviwed
	Score       Score     // score
}

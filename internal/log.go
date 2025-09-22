package internal

import (
	"encoding/csv"
	"time"
	"fmt"
)

// Log records what what time a card was reviewed and what score it had
type Log struct {
	ReviewedAt  time.Time
	Score       Score
}

func (l *Log) Record(s Score) {
	l.ReviewedAt = time.Now()
	l.Score = s
}

func WriteLog(d *Deck) error {
	var logs [][]string
	for i := range d.Cards {
		if !d.Cards[i].Log.ReviewedAt.IsZero() {
			logs = append(logs, []string {
				fmt.Sprintf("%d", d.Cards[i].Meta.Hash),
				fmt.Sprintf("%d", d.Cards[i].Log.ReviewedAt.UnixMilli()),
				fmt.Sprintf("%d", d.Cards[i].Log.Score),
			})
		}
	}
	logWriter, err := d.LogWriter()
	if err != nil {
		return err
	}
	defer logWriter.Close()
	csvWriter := csv.NewWriter(logWriter)
  return csvWriter.WriteAll(logs)
}

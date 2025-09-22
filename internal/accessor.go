package internal

import (
	"io"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// DeckAccessor abstract IO operations around Deck handling.
type DeckAccessor interface {
	DeckName() string
	CardsReader() (io.ReadCloser, error)
	MetaReader() (io.ReadCloser, error)
	MetaWriter() (io.WriteCloser, error)
	LogWriter() (io.WriteCloser, error)
}

type fileAccessor struct {
	filename string
}

func newFileDeckAccessor(filename string) DeckAccessor {
	return &fileAccessor{filename}
}

func (f *fileAccessor) hiddenFile(extension string) string {
	base := filepath.Base(f.filename)
	base = "." + base + extension
	dir := filepath.Dir(f.filename)
	return filepath.Join(dir, base)
}

func (f *fileAccessor) logFile() string {
  return f.hiddenFile(".log")
}

func (f *fileAccessor) metaFile() string {
  return f.hiddenFile(".db")
}

func (f *fileAccessor) CardsReader() (io.ReadCloser, error) {
	return os.Open(f.filename)
}

func (f *fileAccessor) MetaReader() (io.ReadCloser, error) {
	return os.Open(f.metaFile())
}

func (f *fileAccessor) MetaWriter() (io.WriteCloser, error) {
	return os.Create(f.metaFile())
}

func (f *fileAccessor) LogWriter() (io.WriteCloser, error) {
	file, err := os.OpenFile(f.logFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open or create log file %s: %w", f.logFile(), err)
	}
	return file, nil
}

func (f *fileAccessor) DeckName() string {
	return path.Base(f.filename)
}

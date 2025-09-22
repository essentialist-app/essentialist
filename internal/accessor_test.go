package internal

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileAccessor_DeckName(t *testing.T) {
	accessor := newFileDeckAccessor("/some/path/to/my_deck.md")
	expected := "my_deck.md"
	if got := accessor.DeckName(); got != expected {
		t.Errorf("DeckName() = %q, want %q", got, expected)
	}
}

func TestFileAccessor_logFile(t *testing.T) {
	fa := &fileAccessor{filename: "/path/to/my_deck.md"}
	expected := "/path/to/.my_deck.md.log"
	if got := fa.logFile(); got != expected {
		t.Errorf("logFile() = %q, want %q", got, expected)
	}
}

func TestFileAccessor_metaFile(t *testing.T) {
	fa := &fileAccessor{filename: "/path/to/my_deck.md"}
	expected := "/path/to/.my_deck.md.db"
	if got := fa.metaFile(); got != expected {
		t.Errorf("metaFile() = %q, want %q", got, expected)
	}
}

func TestFileAccessor_FileOperations(t *testing.T) {
	dir := t.TempDir()
	deckPath := filepath.Join(dir, "test_deck.md")
	deckContent := "Card 1\nCard 2"
	if err := os.WriteFile(deckPath, []byte(deckContent), 0644); err != nil {
		t.Fatalf("Failed to create test deck file: %v", err)
	}

	accessor := newFileDeckAccessor(deckPath)

	t.Run("CardsReader", func(t *testing.T) {
		reader, err := accessor.CardsReader()
		if err != nil {
			t.Fatalf("CardsReader() error = %v", err)
		}
		defer reader.Close()
		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read from CardsReader: %v", err)
		}
		if string(content) != deckContent {
			t.Errorf("CardsReader() content = %q, want %q", string(content), deckContent)
		}
	})

	t.Run("MetaWriterAndReader", func(t *testing.T) {
		// Test writer
		metaContent := "some metadata"
		writer, err := accessor.MetaWriter()
		if err != nil {
			t.Fatalf("MetaWriter() error = %v", err)
		}
		if _, err := writer.Write([]byte(metaContent)); err != nil {
			t.Fatalf("Failed to write to MetaWriter: %v", err)
		}
		writer.Close()

		// Test reader
		reader, err := accessor.MetaReader()
		if err != nil {
			t.Fatalf("MetaReader() error = %v", err)
		}
		defer reader.Close()
		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read from MetaReader: %v", err)
		}
		if string(content) != metaContent {
			t.Errorf("MetaReader() content = %q, want %q", string(content), metaContent)
		}
	})

	t.Run("LogWriter", func(t *testing.T) {
		logEntry1 := "first log entry\n"
		logWriter, err := accessor.LogWriter()
		if err != nil {
			t.Fatalf("LogWriter() [1] error = %v", err)
		}
		if _, err := logWriter.Write([]byte(logEntry1)); err != nil {
			t.Fatalf("Failed to write to LogWriter [1]: %v", err)
		}
		logWriter.Close()

		logEntry2 := "second log entry\n"
		logWriter, err = accessor.LogWriter()
		if err != nil {
			t.Fatalf("LogWriter() [2] error = %v", err)
		}
		if _, err := logWriter.Write([]byte(logEntry2)); err != nil {
			t.Fatalf("Failed to write to LogWriter [2]: %v", err)
		}
		logWriter.Close()

		fa := accessor.(*fileAccessor)
		logContent, err := os.ReadFile(fa.logFile())
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		if !strings.Contains(string(logContent), logEntry1) {
			t.Errorf("Log content does not contain the first entry. Got: %q", string(logContent))
		}
		if !strings.Contains(string(logContent), logEntry2) {
			t.Errorf("Log content does not contain the second entry. Got: %q", string(logContent))
		}
	})
}

func TestFileAccessor_NonExistentFiles(t *testing.T) {
	dir := t.TempDir()
	accessor := newFileDeckAccessor(filepath.Join(dir, "non_existent.md"))

	t.Run("CardsReader non-existent", func(t *testing.T) {
		_, err := accessor.CardsReader()
		if !os.IsNotExist(err) {
			t.Errorf("Expected CardsReader to return a 'file does not exist' error, but got %v", err)
		}
	})

	t.Run("MetaReader non-existent", func(t *testing.T) {
		_, err := accessor.MetaReader()
		if !os.IsNotExist(err) {
			t.Errorf("Expected MetaReader to return a 'file does not exist' error, but got %v", err)
		}
	})
}

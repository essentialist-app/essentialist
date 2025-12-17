package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	// Import internal to reuse utility if needed, though for server we might just duplicate logic or share types
)

// We need a way to list decks. Since internal.DeckAccessor is an interface, the server needs to know how to find files.
// We will just scan the directory given.

var dataDir = flag.String("data", ".", "directory containing flashcards")
var staticDir = flag.String("static", "./cmd/essentialist/wasm", "directory containing static web files")

func main() {
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(*staticDir)))

	http.HandleFunc("/api/decks", handleListDecks)
	http.HandleFunc("/api/deck/", handleDeckOperations)

	log.Printf("Serving on :8080 from %s (data) and %s (static)", *dataDir, *staticDir)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type DeckInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func handleListDecks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(*dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var decks []DeckInfo
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			decks = append(decks, DeckInfo{
				Name: e.Name(),
				Path: e.Name(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decks)
}

func handleDeckOperations(w http.ResponseWriter, r *http.Request) {
	// Path: /api/deck/{name}/{operation}
	// Operations: content (GET), db (GET, PUT)

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	deckName := parts[3]
	operation := parts[4]

	// Basic security: prevent traversal
	if strings.Contains(deckName, "..") || strings.Contains(deckName, "/") || strings.Contains(deckName, "\\") {
		http.Error(w, "Invalid deck name", http.StatusBadRequest)
		return
	}

	deckPath := filepath.Join(*dataDir, deckName)
	if _, err := os.Stat(deckPath); os.IsNotExist(err) {
		http.Error(w, "Deck not found", http.StatusNotFound)
		return
	}

	switch operation {
	case "content":
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, deckPath)

	case "db":
		dbPath := filepath.Join(*dataDir, "."+deckName+".db")

		if r.Method == http.MethodGet {
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				// Return 404 is fine, or empty json
				http.Error(w, "DB not found", http.StatusNotFound)
				return
			}
			http.ServeFile(w, r, dbPath)
		} else if r.Method == http.MethodPut {
			f, err := os.Create(dbPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer f.Close()
			_, err = io.Copy(f, r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	default:
		http.Error(w, "Unknown operation", http.StatusBadRequest)
	}
}

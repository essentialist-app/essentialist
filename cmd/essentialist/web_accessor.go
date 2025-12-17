package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/essentialist-app/essentialist/internal"
)

type RemoteDeckAccessor struct {
	Name string
}

func (r *RemoteDeckAccessor) DeckName() string {
	return r.Name
}

func (r *RemoteDeckAccessor) Path() string {
	return r.Name
}

// CardsReader fetches the markdown content
func (r *RemoteDeckAccessor) CardsReader() (io.ReadCloser, error) {
	resp, err := http.Get("/api/deck/" + r.Name + "/content")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to fetch deck content: %s", resp.Status)
	}
	return resp.Body, nil
}

// MetaReader fetches the DB content
func (r *RemoteDeckAccessor) MetaReader() (io.ReadCloser, error) {
	resp, err := http.Get("/api/deck/" + r.Name + "/db")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		// Return error so internal/deck.go calls MetaWriter to create new one?
		// internal/deck.go:53: err = newMetaCards(accessor)
		return nil, errors.New("db not found")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to fetch DB content: %s", resp.Status)
	}
	return resp.Body, nil
}

// MetaWriter needs to collect data and then PUT it.
func (r *RemoteDeckAccessor) MetaWriter() (io.WriteCloser, error) {
	// We return a specialized writer that buffers everything and sends on Close.
	return &remoteByteWriter{deckName: r.Name}, nil
}

type remoteByteWriter struct {
	buf      bytes.Buffer
	deckName string
}

func (rb *remoteByteWriter) Write(p []byte) (int, error) {
	return rb.buf.Write(p)
}

func (rb *remoteByteWriter) Close() error {
	req, err := http.NewRequest(http.MethodPut, "/api/deck/"+rb.deckName+"/db", &rb.buf)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to save DB: %s", resp.Status)
	}
	return nil
}

// Helper to fetch list of decks
type DeckInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func fetchRemoteDecks() ([]internal.DeckAccessor, error) {
	resp, err := http.Get("/api/decks")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list decks: %s", resp.Status)
	}

	var infos []DeckInfo
	if err := json.NewDecoder(resp.Body).Decode(&infos); err != nil {
		return nil, err
	}

	var accessors []internal.DeckAccessor
	for _, info := range infos {
		accessors = append(accessors, &RemoteDeckAccessor{Name: info.Name})
	}
	return accessors, nil
}

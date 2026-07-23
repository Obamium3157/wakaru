package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"wakaru/internal/wakaru"
)

func translateHandler(w *wakaru.Wakaru) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(rw, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.Text == "" {
			http.Error(rw, "text is required", http.StatusBadRequest)
			return
		}

		dispStr, results, err := w.Run(r.Context(), req.Text)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]any{
			"displayString": dispStr,
			"results":       results,
		})
	}
}

func wordHandler(w *wakaru.Wakaru) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		text := r.PathValue("text")
		if text == "" {
			http.Error(rw, "text is required", http.StatusBadRequest)
			return
		}

		posMajor := r.URL.Query().Get("pos")

		entries, err := w.FindWord(r.Context(), text, posMajor)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]any{
			"entries": entries,
		})
	}
}

func spaHandler(distDir string) http.HandlerFunc {
	fileServer := http.FileServer(http.Dir(distDir))
	return func(rw http.ResponseWriter, r *http.Request) {
		path := filepath.Join(distDir, filepath.Clean(r.URL.Path))
		cleanDist := filepath.Clean(distDir)
		if path == cleanDist || strings.HasPrefix(path, cleanDist+string(os.PathSeparator)) {
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(rw, r)
				return
			}
		}
		http.ServeFile(rw, r, filepath.Join(distDir, "index.html"))
	}
}

func ankiHandler(w *wakaru.Wakaru) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req struct {
			DeckName string   `json:"deckName"`
			Front    string   `json:"Front"`
			Back     string   `json:"Back"`
			Tags     []string `json:"tags"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(rw, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.DeckName == "" || req.Front == "" || req.Back == "" {
			http.Error(rw, "deckName, front, back are required", http.StatusBadRequest)
			return
		}

		if err := w.AddBasicNote(
			r.Context(),
			req.DeckName,
			req.Front,
			req.Back,
			req.Tags,
		); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusNoContent)
	}
}

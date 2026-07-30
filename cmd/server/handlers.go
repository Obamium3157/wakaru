package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"wakaru/internal/examples"
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

		flusher, ok := rw.(http.Flusher)
		if !ok {
			http.Error(rw, "streaming not supported", http.StatusInternalServerError)
			return
		}

		var writeMu sync.Mutex
		sendEvent := func(event string, data any) {
			if r.Context().Err() != nil {
				return
			}
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				log.Printf("sendEvent marshal error: %v", err)
				return
			}

			writeMu.Lock()
			defer writeMu.Unlock()
			if _, err := fmt.Fprintf(rw, "event: %s\ndata: %s\n\n", event, jsonBytes); err != nil {
				log.Println(err)
				return
			}
			flusher.Flush()
		}

		rw.Header().Set("Content-Type", "text/event-stream")
		rw.Header().Set("Cache-Control", "no-cache")
		rw.Header().Set("Connection", "keep-alive")
		rw.Header().Set("X-Accel-Buffering", "no")

		err := w.RunStream(r.Context(), req.Text, wakaru.StreamCallbacks{
			OnInit: func(displayString string, results []wakaru.Result) {
				sendEvent("init", map[string]any{
					"displayString": displayString,
					"results":       results,
				})
			},
			OnExamples: func(index int, examples []examples.Example) {
				sendEvent("examples", map[string]any{
					"index":    index,
					"examples": examples,
				})
			},
			OnDone: func() {
				sendEvent("done", map[string]any{})
			},
		})
		if err != nil {
			log.Printf("translate error: %v\n", err)
			return
		}
		log.Println("RunStream finished")
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
			log.Printf("word lookup error: %v", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(rw).Encode(map[string]any{
			"entries": entries,
		}); err != nil {
			log.Printf("word lookup encode error: %v", err)
		}
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
			log.Printf("anki note error: %v", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusNoContent)
	}
}

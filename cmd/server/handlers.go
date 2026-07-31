package main

import (
	"context"
	"encoding/json"
	"errors"
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

type translateRequest struct {
	Text string `json:"text"`
}

func translateHandler(w *wakaru.Wakaru) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		req, err := decodeTranslateRequest(r)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		sse, err := newSSEWriter(rw)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = w.RunStream(r.Context(), req.Text, wakaru.StreamCallbacks{
			OnInit: func(displayString string, results []wakaru.Result) {
				sse.SendEvent(r.Context(), "init", map[string]any{
					"displayString": displayString,
					"results":       results,
				})
			},
			OnExamples: func(index int, e []examples.Example) {
				sse.SendEvent(r.Context(), "examples", map[string]any{
					"index":    index,
					"examples": e,
				})
			},
			OnDone: func() {
				sse.SendEvent(r.Context(), "done", map[string]any{})
			},
		})
		if err != nil {
			log.Printf("translate error: %v\n", err)
			return
		}
		log.Println("RunStream finished")
	}
}

func decodeTranslateRequest(r *http.Request) (translateRequest, error) {
	var req translateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, errors.New("invalid request body")
	}
	if req.Text == "" {
		return req, errors.New("text is required")
	}

	return req, nil
}

type sseWriter struct {
	rw      http.ResponseWriter
	flusher http.Flusher
	writeMu sync.Mutex
}

func newSSEWriter(rw http.ResponseWriter) (*sseWriter, error) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("X-Accel-Buffering", "no")

	return &sseWriter{
		rw:      rw,
		flusher: flusher,
	}, nil
}

func (w *sseWriter) SendEvent(ctx context.Context, event string, data any) {
	if ctx.Err() != nil {
		return
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("sendEvent marshal error: %v", err)
		return
	}

	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	if _, err := fmt.Fprintf(w.rw, "event: %s\ndata: %s\n\n", event, jsonBytes); err != nil {
		log.Println(err)
		return
	}

	w.flusher.Flush()
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

type aiExamplesRequest struct {
	Word string `json:"word"`
}

func aiExamplesHandler(w *wakaru.Wakaru) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		req, err := decodeAIExamplesRequest(r)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		examples, err := w.GenerateAIExamples(r.Context(), req.Word)
		if err != nil {
			log.Printf("ai examples error: %v", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(rw).Encode(map[string]any{
			"examples": examples,
		}); err != nil {
			log.Printf("ai examples encode error: %v", err)
		}
	}
}

func decodeAIExamplesRequest(r *http.Request) (aiExamplesRequest, error) {
	var req aiExamplesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, errors.New("invalid request body")
	}
	if req.Word == "" {
		return req, errors.New("word is required")
	}

	return req, nil
}

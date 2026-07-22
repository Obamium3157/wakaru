package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"

	"wakaru/internal/wakaru"

	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	mustLoadEnv()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	w, err := wakaru.NewWakaru(ctx, "sqlite3", os.Getenv("DB_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/translate", translateHandler(w))
	mux.HandleFunc("GET /api/word/{text}", wordHandler(w))
	mux.HandleFunc("GET /{path...}", spaHandler("frontend/dist"))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("listening on :8080")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func mustLoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
}

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
		path := distDir + r.URL.Path
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(rw, r)
			return
		}
		http.ServeFile(rw, r, distDir+"/index.html")
	}
}

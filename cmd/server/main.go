package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"wakaru/internal/wakaru"

	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	mustLoadEnv()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ankiPort, err := strconv.Atoi(os.Getenv("ANKI_PORT"))
	if err != nil {
		log.Fatalf("failed parsing anki port: %v", err)
	}

	w, err := wakaru.NewWakaru(
		ctx,
		"sqlite3",
		os.Getenv("DB_PATH"),
		ankiPort,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	mux := initMux(w)

	port, err := getPort()
	if err != nil {
		log.Fatal(err)
	}

	srv := initServer(port, mux)

	go func() {
		<-ctx.Done()
		log.Println("shutting down...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("listening on %s", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func mustLoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
}

func getPort() (string, error) {
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		return "", err
	}
	if ok := validatePort(port); !ok {
		return "", fmt.Errorf("server port should be in range [1, 65535], got %d", port)
	}

	return fmt.Sprintf(":%d", port), nil
}

func validatePort(port int) bool {
	return 1 <= port && port <= 65535
}

func initMux(w *wakaru.Wakaru) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/translate", translateHandler(w))
	mux.HandleFunc("GET /api/word/{text}", wordHandler(w))
	mux.HandleFunc("POST /api/anki/note", ankiHandler(w))
	mux.HandleFunc("GET /{path...}", spaHandler("frontend/dist"))

	return mux
}

func initServer(port string, handler *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:    port,
		Handler: handler,
	}
}

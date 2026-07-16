.PHONY: setup build run dev clean

setup:
	cd frontend && npm install
	go mod download

build:
	cd frontend && npm run build
	go build -o wakaru ./cmd/server

run: build
	./wakaru

dev:
	go run ./cmd/server &
	cd frontend && npm run dev

clean:
	rm -f wakaru
	rm -rf frontend/dist frontend/node_modules

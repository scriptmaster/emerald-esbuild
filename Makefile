all: demo

demo: build
	emerald-esbuild

build:
	go build -ldflags "-s -w" -o /usr/local/bin/emerald-esbuild .

run:
	go run .

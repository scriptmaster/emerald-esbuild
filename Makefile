all: demo

demo: build
	emerald-esbuild

build:
	go build -ldflags "-s -w" -o /usr/local/bin/emerald-esbuild .

run:
	go run .

test:
	go test ./...

release:
	git tab v1.0.0
	git push origin v1.0.0

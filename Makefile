build:
	go build -o altituder ./cmd/altituder

test:
	go test ./...

run:
	go run ./cmd/altituder

air-bin:
	dlv exec --headless --continue --listen :2345 --accept-multiclient --log --log-dest=/app/.dev/devel-debug.log ./.dev/main serve
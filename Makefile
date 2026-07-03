.PHONY: test test-integration build docker run tidy

BINARY := bin/usrsvr

test:
	CGO_ENABLED=0 go test ./...

test-integration:
	go test -tags=integration ./test/integration/...

build:
	go build -o $(BINARY) ./cmd/usrsvr

docker:
	docker build -f deploy/docker/Dockerfile -t hi-im-usrsvr:latest .

run: build
	./$(BINARY)

tidy:
	go mod tidy

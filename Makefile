BIN=./bin/previewer

build:
	go build -v -o $(BIN) ./cmd

build-linux:
	GOOS=linux go build -o $(BIN)-linux ./cmd

build-image:
	docker build -t proxy-previewer .

run-container: build-linux build-image
	docker run -it --rm -p 8081:8081 proxy-previewer

run: build-linux build-image
	docker-compose up

test:
	go clean -cache
	go test -v ./...
	go test -race -count 10 ./internal/cache

run-img-srv:
	docker run --name img-srv --rm -v $(shell pwd)/images/www:/usr/share/nginx/html -p 8082:80 -d nginx:alpine

kill-img-srv:
	docker stop img-srv

lint:
	golangci-lint run ./...

.PHONY: test

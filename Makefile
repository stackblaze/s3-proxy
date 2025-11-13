.PHONY: build start stop test

build:
	go build -o s3-proxy .

test:
	go test -v ./...

start:
	./s3-proxy -port 8080

stop:
	pkill -f s3-proxy || true


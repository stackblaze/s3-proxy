.PHONY: build start stop

build:
	go build -o s3-proxy .

start:
	./s3-proxy -port 8080

stop:
	pkill -f s3-proxy || true


.PHONY: build install clean test

build:
	go build -o nutshell ./cmd/nutshell/

install:
	go install ./cmd/nutshell/

clean:
	rm -f nutshell

test:
	go test ./pkg/nutshell/...

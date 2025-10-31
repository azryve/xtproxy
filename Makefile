.PHONY: all generate clean xtproxy tests lint format

all: xtproxy

clean:
	[ -e xtproxy ] && rm xtproxy

xtproxy:
	go build ./cmd/xtproxy/

test:
	go test -v ./...

lint:
	@go vet ./...
	@gofmt -s -d .
	@test -z $$(gofmt -s -l .)

format:
	@gofmt -s -w .

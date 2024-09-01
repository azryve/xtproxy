.PHONY: all generate clean xtproxy tests

all: xtproxy

clean:
	[ -e xtproxy ] && rm xtproxy

xtproxy:
	go build ./cmd/xtproxy/

tests:
	go test ./pkg/...

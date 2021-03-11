.PHONY: all clean help install

SRC := $(shell find . -name '*.go')
BIN := $(subst .go,,$(wildcard cmd/*.go))


all: $(BIN)	# build all binary

clean:		# clean-up the environment
	rm -f $(BIN)

help:		# show this message
	@printf "Usage: make [OPTION]\n"
	@printf "\n"
	@perl -nle 'print $$& if m{^[\w-]+:.*?#.*$$}' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?#"} {printf "    %-18s %s\n", $$1, $$2}'

linter: $(SRC)
	gofmt -w -s $^
	go test -cover -failfast -timeout 2s ./...
	go test -bench=. -cover -failfast -benchmem

$(BIN): linter

%: %.go
	go build  -ldflags="-s -w" -o $@ $<

PREFIX := /usr/local/bin
install: $(BIN)	# install the binary to the PREFIX
	install -m755 $^ $(PREFIX)/

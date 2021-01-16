.PHONY: all clean help install

SRC := $(shell find . -name '*.go')
BIN := $(subst .go,,$(wildcard cmd/*.go))


all: $(BIN) linter	# build all binary

clean:		# clean-up the environment
	rm -f $(BIN)

help:		# show this message
	@printf "Usage: make [OPTION]\n"
	@printf "\n"
	@perl -nle 'print $$& if m{^[\w-]+:.*?#.*$$}' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?#"} {printf "    %-18s %s\n", $$1, $$2}'

PREFIX := /usr/local/bin
install: $(BIN)	# install the binary to the PREFIX
	install -m755 $^ $(PREFIX)/

GO      := go
GOFMT   := $(GO)fmt -w -s
GOFLAG  := -ldflags="-s -w"
GOTEST  := $(GO) test -cover -failfast -timeout 2s
GOBENCH := $(GO) test -bench=. -cover -failfast -benchmem

linter: .benchmark
	$(GOFMT) $(shell find . -name '*.go')
	$(GOTEST) ./...

.benchmark: $(SRC)
	touch $@
	$(GOBENCH)

$(BIN): linter

%: %.go
	$(GO) build $(GOFLAG) -o $@ $<

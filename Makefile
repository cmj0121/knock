SRC := $(wildcard *.go)
BIN := knock

.PHONY: all clean test help

all: $(BIN)			# default action
	@pre-commit install --install-hooks
	@git config commit.template .git-commit-template

clean: 				# clean-up environment
	@rm -f $(BIN)

test: 				# run test
	gofmt -w -s $(SRC)
	go test -cover -failfast -timeout 2s ./...

help:	# show this message
	@printf "Usage: make [OPTION]\n"
	@printf "\n"
	@perl -nle 'print $$& if m{^[\w-]+:.*?#.*$$}' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?#"} {printf "    %-18s %s\n", $$1, $$2}'

$(BIN): $(SRC) test
	go build -ldflags="-s -w" -o $@ $<

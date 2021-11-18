SRC := $(wildcard *.go) $(wildcard */*.go)
BIN := $(subst .go,,$(wildcard cmd/*.go))

.PHONY: all clean test run build upgrade help

all: build		# default action
	@pre-commit install --install-hooks >/dev/null
	@git config commit.template .git-commit-template

clean: 			# clean-up environment
	@rm -f $(BIN)

test: 			# run test
	gofmt -w -s $(SRC)
	go test -cover -failfast -timeout 2s ./...

run:			# run in the local environment


build: $(BIN)	# build the binary

upgrade:		# upgrade all the necessary packages
	pre-commit autoupdate

help:			# show this message
	@printf "Usage: make [OPTION]\n"
	@printf "\n"
	@perl -nle 'print $$& if m{^[\w-]+:.*?#.*$$}' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?#"} {printf "    %-18s %s\n", $$1, $$2}'

%: %.go
	GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="-s -w" -o $@ $<

SHELL := /bin/bash
TMPDIR := $(if $(TMPDIR),$(TMPDIR),"/tmp/")
GOPATH := $(shell go env GOPATH)

tulpa_bin := $(GOPATH)/bin/tulpa
gofiles := $(wildcard *.go **/*.go **/**/*.go **/**/**/*.go)

gocoverutil := $(GOPATH)/bin/gocoverutil
staticcheck := $(GOPATH)/bin/staticcheck
golangcilint := $(GOPATH)/bin/golangci-lint
gomodoutdated := $(GOPATH)/bin/go-mod-outdated

all: build

.PHONY: build
build: $(tulpa_bin)

$(tulpa_bin): $(gofiles)
	GO111MODULE=on go install ./...

.PHONY: test
test:
	GO111MODULE=on go test -cover -race ./...

.PHONY: test.lint
test.lint: $(staticcheck)
	staticcheck -f stylish ./...

.PHONY: test.cover
test.cover: $(gocoverutil)
	gocoverutil -coverprofile=cover.out test -covermode=count ./... \
		2> >(grep -v "no packages being tested depend on matches for pattern" 1>&2) \
		| sed -e 's/of statements in .*/of statements/'
	# go tool cover -func cov.out
	@echo -n "total: "; go tool cover -func=cov.out | tail -n 1 | sed -e 's/\((statements)\|total:\)//g' | tr -s "[:space:]"

.PHONY: test.outdated
test.outdated: $(gomodoutdated)
	GO111MODULE=on go list -u -m -json all | go-mod-outdated -direct

.PHONY: release.dryrun
release.dryrun:
	goreleaser --snapshot --skip-publish --rm-dist

$(gocoverutil):
	GO111MODULE=off go get github.com/AlekSi/gocoverutil

$(staticcheck):
	cd $(TMPDIR) && GO111MODULE=on go get honnef.co/go/tools/cmd/staticcheck@2019.2.3

$(golangcilint):
	cd $(TMPDIR) && GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.18.0

$(gomodoutdated):
	GO111MODULE=off go get github.com/psampaz/go-mod-outdated


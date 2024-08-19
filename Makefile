PREFIX := github.com/macstadium/packer-plugin-macstadium-orka
VERSION := $(shell git describe --tags --candidates=1 --dirty 2>/dev/null || echo "dev")
FLAGS := -X version/version.Version=$(VERSION)
BIN := packer-plugin-macstadium-orka
SOURCES := $(shell find . -name '*.go')
GOOS ?= darwin
TEST ?= builder/orka/*.go

.PHONY: all test testacc install-gen-deps generate build install \
	packer-build-example packer-build-example-non-debug plugin-check \
	fresh rebuild clean
all: rebuild

$(BIN): $(SOURCES)
	go build -ldflags="$(FLAGS)" -o $(BIN) $(PREFIX)

test:
	go test -v $(TEST)

testacc:
	PACKER_ACC=1 go test -count 1 -v $(TEST) -timeout=180m

install-gen-deps:
	GOBIN=$(shell pwd) go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@latest

generate: install-gen-deps
	PATH="$(shell pwd):${PATH}" go generate builder/orka/config.go

build: generate $(BIN)

install: $(BIN)
	@mkdir -p ~/.packer.d/plugins/
	@mv $(BIN) ~/.packer.d/plugins/

packer-build-example:
	PACKER_LOG=1 packer build -on-error=ask examples/orka.pkr.hcl

packer-build-example-non-debug:
	packer build examples/orka.pkr.hcl

plugin-check: build
	PATH="$(shell pwd):${PATH}" packer-sdc plugin-check $(BIN)

fresh: clean build install packer-build-example-non-debug clean

rebuild: build install clean

clean:
	@rm -f $(BIN)
	@rm -f packer-sdc

.PHONY: fmt
fmt:
	@echo "[fmt] Format go project..."
	@gofmt -s -w . 2>&1
	@echo "------------------------------------[Done]"

.PHONY: tidy
tidy:
	@echo "[tidy] Check for unused modules..."
	@go mod tidy
	@echo "------------------------------------[Done]"

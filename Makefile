PREFIX := github.com/andrewfarley/packer-builder-macstadium-orka
VERSION := $(shell git describe --tags --candidates=1 --dirty 2>/dev/null || echo "dev")
FLAGS := -X main.Version=$(VERSION)
BIN := packer-builder-macstadium-orka
SOURCES := $(shell find . -name '*.go')
GOOS ?= darwin

.PHONY: clean

test:
	go test -v builder/orka/*.go

build: $(BIN)

$(BIN): $(SOURCES)
	GOBIN=$(shell pwd) go install github.com/hashicorp/packer/cmd/mapstructure-to-hcl2
	PATH=$(shell pwd):${PATH} go generate builder/orka/config.go
	go build -ldflags="$(FLAGS)" -o $(BIN) $(PREFIX)

install: $(BIN)
	mkdir -p ~/.packer.d/plugins/
	cp $(BIN) ~/.packer.d/plugins/

packer-build-example:
	PACKER_LOG=1 packer build -on-error=ask examples/macos-catalina.json

packer-build-example-non-debug:
	packer build examples/macos-catalina.json

fresh: clean build install packer-build-example

rebuild: clean build install

clean:
	rm -f $(BIN)

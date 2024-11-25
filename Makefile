VERSION := $(shell git describe --tags --always --dirty)
BINARY_NAME := tsky

LDFLAGS := -X 'main.Version=$(VERSION)'

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o "$(BINARY_NAME)"

clean:
	rm -f myapp

.PHONY: all build clean
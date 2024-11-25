VERSION := $(shell git describe --tags --always --dirty)
BINARY_NAME := tsky

LDFLAGS := -X 'main.Version=$(VERSION)'

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o "$(BINARY_NAME)"

clean:
	rm -f "$(BINARY_NAME)"

.PHONY: all build clean
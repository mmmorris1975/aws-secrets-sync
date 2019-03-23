EXE  := secrets-sync
VER  := $(shell git describe --tags)
PATH := build:$(PATH)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: darwin linux windows release clean dist-clean test

$(EXE): go.mod *.go
	go build -v -ldflags '-X main.Version=$(VER)' -o $@

darwin linux:
	GOOS=$@ go build -ldflags '-X main.Version=$(VER)' -o $(EXE)-$(VER)-$@-$(GOARCH)

# $(shell go env GOEXE) is evaluated in the context of the Makefile host (before GOOS is evaluated), so hard-code .exe
windows:
	GOOS=$@ go build -ldflags '-X main.Version=$(VER)' -o $(EXE)-$(VER)-$@-$(GOARCH).exe

clean:
	rm -f $(EXE) $(EXE)-*-*-*

dist-clean: clean
	rm -f go.sum

docker: linux
	docker build . -t $(EXE):$(VER)

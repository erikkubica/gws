.PHONY: all build test install uninstall validate lint clean

all: build

build:
	go build -o bin/gws ./cmd/gws

test:
	go test -v ./...

install:
	./install.sh

uninstall:
	./install.sh --uninstall

validate:
	agy plugin validate plugin/antigravity

lint:
	biome check plugin/

clean:
	rm -rf bin/gws bin/gmcp

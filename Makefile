.PHONY: build run test clean

CGO_CFLAGS ?= -I$(HOME)/.local/usr/include
CGO_LDFLAGS ?= -L$(HOME)/.local/usr/lib/x86_64-linux-gnu

build:
	CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go build -tags x11 -o bin/cvchecker ./cmd/cvchecker

run: build
	./bin/cvchecker

test:
	CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go test -tags x11 -v ./...

clean:
	rm -rf bin/

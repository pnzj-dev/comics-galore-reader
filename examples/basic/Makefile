.PHONY: all build clean serve

GOROOT ?= $(shell go env GOROOT)
WASM_EXEC = $(GOROOT)/lib/wasm/wasm_exec.js

all: build

# Build the WASM binary.
build:
	GOOS=js GOARCH=wasm go build -o cg-reader-wasm.wasm .
	@echo "Built cg-reader-wasm.wasm"

# Run templ generate to produce reader_templ.go from reader.templ.
templ:
	cd ../.. && templ generate

# Copy wasm_exec.js from the Go installation.
wasm_exec.js:
	cp "$(WASM_EXEC)" wasm_exec.js

# Copy all needed files for serving.
prepare: build wasm_exec.js
	@echo "All files ready in examples/basic/"

# Start a local dev server.
serve: prepare
	@echo "Serving at http://localhost:8080"
	python3 -m http.server 8080

# Clean build artifacts.
clean:
	rm -f cg-reader-wasm.wasm wasm_exec.js

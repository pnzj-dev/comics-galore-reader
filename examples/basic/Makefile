.PHONY: all build clean serve rebuild templ wasm_exec.js

GOROOT     ?= $(shell go env GOROOT)
WASM_EXEC  = $(GOROOT)/lib/wasm/wasm_exec.js
WASM_BIN   = cg-reader-wasm.wasm
PORT       ?= 8080

all: build

templ:
	cd ../.. && templ generate

build: templ
	@echo "Building $(WASM_BIN)..."
	GOOS=js GOARCH=wasm go build -o $(WASM_BIN) .
	@ls -lh $(WASM_BIN)

wasm_exec.js:
	@if [ -f "$(WASM_EXEC)" ]; then \
		cp "$(WASM_EXEC)" wasm_exec.js && echo "Copied wasm_exec.js"; \
	else \
		echo "ERROR: wasm_exec.js not found at $(WASM_EXEC)"; exit 1; \
	fi

prepare: build wasm_exec.js
	@echo "Ready at http://localhost:$(PORT)"

serve: prepare
	@python3 serve.py $(PORT) "$$(pwd)"

rebuild:
	@lsof -ti:$(PORT) | xargs kill -9 2>/dev/null || true
	@$(MAKE) serve

clean:
	rm -f $(WASM_BIN) wasm_exec.js

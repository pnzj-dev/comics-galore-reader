.PHONY: all build clean serve rebuild templ css static

GOROOT     ?= $(shell go env GOROOT)
WASM_EXEC  = $(GOROOT)/lib/wasm/wasm_exec.js
WASM_BIN   = cg-reader-wasm.wasm
PORT       ?= 8080

all: build

templ:
	cd ../.. && templ generate

css:
	@./tailwindcss -i ../../static/input.css -o tailwind.css 2>&1
	@cp tailwind.css ../../static/tailwind.css
	@echo "Tailwind CSS compiled."

static: wasm_exec.js
	@cp ../../static/cg-reader-wasm.js cg-reader-wasm.js
	@echo "Static assets copied."

build: templ css static
	@echo "Building $(WASM_BIN)..."
	GOOS=js GOARCH=wasm go build -o $(WASM_BIN) .
	@ls -lh $(WASM_BIN)

wasm_exec.js:
	@if [ -f "$(WASM_EXEC)" ]; then \
		cp "$(WASM_EXEC)" wasm_exec.js && \
		cp "$(WASM_EXEC)" ../../static/wasm_exec.js && \
		echo "Copied wasm_exec.js"; \
	else \
		echo "ERROR: wasm_exec.js not found at $(WASM_EXEC)"; exit 1; \
	fi

prepare: build
	@echo "Ready at http://localhost:$(PORT)"

serve: prepare
	@python3 serve.py $(PORT) "$$(pwd)"

rebuild:
	@lsof -ti:$(PORT) | xargs kill -9 2>/dev/null || true
	@$(MAKE) serve

clean:
	rm -f $(WASM_BIN) wasm_exec.js tailwind.css cg-reader-wasm.js

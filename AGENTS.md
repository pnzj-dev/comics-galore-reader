**Project**: Go WASM Comic Reader Library (`comicwasm`)

**Goal**  
Build a pure-Go library that can be compiled to WebAssembly for reading CBZ (ZIP) and CBR (RAR) comic archives in the browser, with password support, IndexedDB caching, runtime configuration, and excellent integration with Templ-based UIs.

## Core Requirements

- Support **CBZ** via `archive/zip` (stdlib).
- Support **CBR** via pure-Go RAR decoder (`nwaples/rardecode/v2`).
- **Password Protection**:
  - Retrieve password from a configurable URL.
  - Provide a **default password URL** that can be overridden at runtime.
  - Support automatic fetch + manual fallback.
- **Caching**: Persistent Browser IndexedDB (full archives + individual pages).
- Lazy extraction and on-demand image decoding.
- Natural page sorting.
- Metadata extraction (`ComicInfo.xml`).

## Reader Features

- Basic page navigation (Next, Previous, Go to page).
- Display **"Current Page / Total Pages"**.
- Display **archive title** (from `ComicInfo.xml` when available, otherwise filename).
- Runtime override of password URL.

## Architecture & Package Structure
```

comics-galore-reader/
├── comic.go          # Main public API (ComicReader)
├── archive.go
├── password.go       # Default URL + override logic
├── cache.go          # IndexedDB
├── metadata.go       # ComicInfo.xml parsing
├── rar.go
├── zip.go
├── image.go
├── sort.go
├── js.go             # syscall/js bindings
├── go.mod
├── AGENTS.md
└── examples/
└── wasm/         # index.html + interop

```
## Skills & Expertise Required

Expert Go + WASM engineer skilled in:
- `syscall/js` interop.
- Password flows with runtime overrides.
- IndexedDB via pure-Go bindings.
- Comic metadata handling.
- Clean APIs for Templ / HTMX frontends.

## Configuration

- **Password URL**: Default value (e.g. `/api/comic-password`) + `WithPasswordURL()` option + `SetPasswordURL()` method for runtime override.

## Public API (Go → JS)

```go
type ComicReader struct { ... }

func NewComicReader(opts ...Option) *ComicReader
func WithPasswordURL(url string) Option
func WithCacheEnabled(bool) Option

func (r *ComicReader) OpenArchive(data []byte, filename string) error
func (r *ComicReader) GetPage(page int) ([]byte, error) // 0-based
func (r *ComicReader) Next() bool
func (r *ComicReader) Prev() bool
func (r *ComicReader) SetCurrentPage(page int) error
func (r *ComicReader) PageCount() int
func (r *ComicReader) CurrentPage() int
func (r *ComicReader) Title() string
func (r *ComicReader) Filename() string
```

## UI Layer - Templ Compatibility

**Templ Support**: Fully compatible and **strongly recommended** using a **Hybrid architecture**.

### Recommended Setup

- Use `github.com/a-h/templ` for the main UI (shell, controls, dialogs, cache manager).
- Use the Go WASM library as the **comic engine** (parsing, navigation, caching).
- Bridge via clean JavaScript interop.

### Templ + HTMX Integration Guidelines

- Render the reader shell with Templ.
- Use HTMX for dynamic interactions where possible.
- Call WASM functions for heavy operations (open archive, get page, navigation).
- Update DOM elements (title, page counter) after WASM calls.

## Caching Strategy (IndexedDB)

- Cache full archive after successful open.
- Cache extracted pages for fast navigation.
- Store metadata and reading progress.
- LRU eviction + user cache controls.

## Development Priorities

1. Core reader with password + runtime URL override.
2. Navigation + metadata (title, page counter).
3. IndexedDB caching.
4. Clean JS bindings.
5. Templ + HTMX demo.

**Always prioritize** clean JS API, good defaults, and minimal memory usage.

**Never** use CGO.

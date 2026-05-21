package cgreaderwasm

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
)

// ComicReader is the main public API for reading comic archives.
// It manages archive opening, page navigation, metadata, and caching.
type ComicReader struct {
	mu sync.RWMutex

	archive     Archive
	currentPage int
	filename    string

	// options
	passwordURL  string
	cacheEnabled bool
	cache        *Cache

	// runtime state
	password  string
	pageCache map[int][]byte // in-memory page cache (current + adjacent)
}

// Option configures a ComicReader.
type Option func(*ComicReader)

// WithPasswordURL sets the URL to fetch the archive password from.
// An empty string means no password is needed (unencrypted archives).
func WithPasswordURL(url string) Option {
	return func(r *ComicReader) {
		r.passwordURL = url
	}
}

// WithCacheEnabled enables or disables IndexedDB caching.
func WithCacheEnabled(enabled bool) Option {
	return func(r *ComicReader) {
		r.cacheEnabled = enabled
	}
}

// New creates a new ComicReader with the given options.
func New(opts ...Option) *ComicReader {
	r := &ComicReader{
		currentPage: 0,
		pageCache:   make(map[int][]byte),
		cacheEnabled: true,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// OpenArchive opens a comic archive from raw bytes and a filename.
// It handles format detection and password resolution automatically:
//   - If passwordURL is set, it fetches the password from that URL.
//   - If passwordURL is empty, it attempts to open without a password.
//     If the archive is encrypted, it returns a clear error.
func (r *ComicReader) OpenArchive(data []byte, filename string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Close any previously open archive.
	if r.archive != nil {
		r.archive.Close()
	}

	r.filename = filename
	r.currentPage = 0
	r.pageCache = make(map[int][]byte)

	// Resolve password.
	password, err := r.resolvePassword()
	if err != nil {
		return fmt.Errorf("password resolution: %w", err)
	}

	// Detect format and create the appropriate archive.
	format := DetectFormat(filename)
	archive, err := newArchive(format)
	if err != nil {
		return fmt.Errorf("creating archive reader: %w", err)
	}

	// Try opening with password (may be empty).
	if err := archive.Open(data, filename, password); err != nil {
		// If we didn't provide a password but the archive needs one,
		// give a clear error rather than a cryptic decompression failure.
		if password == "" && isEncryptionError(err) {
			return errors.New("archive is password-protected but no password URL was provided")
		}
		return fmt.Errorf("opening archive: %w", err)
	}

	r.archive = archive

	// Initialize cache if enabled.
	if r.cacheEnabled && r.cache == nil {
		r.cache = NewCache()
	}

	return nil
}

// resolvePassword handles the password resolution logic.
func (r *ComicReader) resolvePassword() (string, error) {
	if r.passwordURL == "" {
		return "", nil // no password needed
	}
	// If we already have a cached password, return it.
	if r.password != "" {
		return r.password, nil
	}
	// Fetch password from URL.
	password, err := FetchPassword(r.passwordURL)
	if err != nil {
		return "", fmt.Errorf("fetching password from %q: %w", r.passwordURL, err)
	}
	r.password = password
	return password, nil
}

// isEncryptionError returns true if the error indicates the archive is encrypted.
func isEncryptionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Check for common encryption/CRC error patterns.
	return containsAny(msg,
		"encrypted", "password", "CRC", "checksum",
		"wrong password", "authentication",
		"invalid compressed data",
	)
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// SetPasswordURL overrides the password URL at runtime.
func (r *ComicReader) SetPasswordURL(url string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.passwordURL = url
	r.password = "" // clear cached password so it's re-fetched next open
}

// ---- Page Navigation ----

// PageCount returns the total number of pages. Returns 0 if no archive is open.
func (r *ComicReader) PageCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.archive == nil {
		return 0
	}
	return r.archive.PageCount()
}

// CurrentPage returns the current 0-based page index.
func (r *ComicReader) CurrentPage() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentPage
}

// SetCurrentPage sets the current page (0-based).
func (r *ComicReader) SetCurrentPage(page int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.archive == nil {
		return errors.New("no archive open")
	}
	count := r.archive.PageCount()
	if count == 0 {
		return errors.New("archive contains no pages")
	}
	if page < 0 || page >= count {
		return fmt.Errorf("page %d out of range [0, %d)", page, count)
	}
	r.currentPage = page
	return nil
}

// Next advances to the next page. Returns false if already at the last page.
func (r *ComicReader) Next() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.archive == nil {
		return false
	}
	if r.currentPage >= r.archive.PageCount()-1 {
		return false
	}
	r.currentPage++
	return true
}

// Prev moves to the previous page. Returns false if already at the first page.
func (r *ComicReader) Prev() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.currentPage <= 0 {
		return false
	}
	r.currentPage--
	return true
}

// ---- Page Retrieval ----

// GetPage returns the raw image bytes for a 0-based page index.
func (r *ComicReader) GetPage(page int) ([]byte, error) {
	r.mu.RLock()
	arch := r.archive
	r.mu.RUnlock()

	if arch == nil {
		return nil, errors.New("no archive open")
	}

	// Check in-memory cache first.
	r.mu.Lock()
	if data, ok := r.pageCache[page]; ok {
		r.mu.Unlock()
		return data, nil
	}
	r.mu.Unlock()

	// Fetch from archive.
	data, err := arch.GetPage(page)
	if err != nil {
		return nil, err
	}

	// Cache in memory.
	r.mu.Lock()
	r.pageCache[page] = data
	// Keep only current page ± 2 in memory to limit memory usage.
	for p := range r.pageCache {
		if p < page-2 || p > page+2 {
			delete(r.pageCache, p)
		}
	}
	r.mu.Unlock()

	return data, nil
}

// GetCurrentPage returns the image bytes for the current page.
func (r *ComicReader) GetCurrentPage() ([]byte, error) {
	return r.GetPage(r.CurrentPage())
}

// ---- Metadata ----

// Title returns the archive title (from ComicInfo.xml or filename fallback).
func (r *ComicReader) Title() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.archive == nil {
		return ""
	}
	title := r.archive.Title()
	if title == "" {
		return r.filename
	}
	return title
}

// Filename returns the original filename of the archive.
func (r *ComicReader) Filename() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.filename
}

// ---- Format-Specific Archive Factory ----

func newArchive(format ArchiveFormat) (Archive, error) {
	switch format {
	case FormatCBZ:
		return &ZipArchive{}, nil
	case FormatCBR:
		return &RarArchive{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %d", format)
	}
}

// ---- Sorting ----

// SortPages sorts page entries using natural ordering (1, 2, ..., 10, 11).
func SortPages(entries []PageEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return NaturalLess(entries[i].Name, entries[j].Name)
	})
}

// Ensure io.Reader is referenced.
var _ io.Reader = nil

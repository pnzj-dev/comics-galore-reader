package cgreaderwasm

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// ZipArchive implements Archive for CBZ files (ZIP format).
type ZipArchive struct {
	filename  string
	title     string
	pages     []PageEntry
	reader    *zip.Reader
	pageCache map[int][]byte
}

func (z *ZipArchive) Open(data []byte, filename string, password string) error {
	z.filename = filename
	z.pages = nil
	z.pageCache = make(map[int][]byte)

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	z.reader = reader

	// Extract page entries (image files only) and metadata.
	var pageEntries []PageEntry
	for i, f := range reader.File {
		name := f.Name
		// Skip directories and non-image files.
		if f.FileInfo().IsDir() {
			continue
		}
		lower := strings.ToLower(name)
		if isImageFile(lower) {
			pageEntries = append(pageEntries, PageEntry{
				Index: i,
				Name:  name,
				Size:  int64(f.UncompressedSize64),
			})
		} else if strings.ToLower(name) == "comicinfo.xml" {
			rc, err := f.Open()
			if err == nil {
				xmlData, _ := io.ReadAll(rc)
				rc.Close()
				z.title = ParseComicInfoTitle(xmlData)
			}
		}
	}

	// Sort pages naturally.
	SortPages(pageEntries)
	// Re-index.
	for i := range pageEntries {
		pageEntries[i].Index = i
	}
	z.pages = pageEntries

	return nil
}

func (z *ZipArchive) PageCount() int {
	return len(z.pages)
}

func (z *ZipArchive) GetPage(page int) ([]byte, error) {
	if page < 0 || page >= len(z.pages) {
		return nil, fmt.Errorf("page %d out of range [0, %d)", page, len(z.pages))
	}

	entry := z.pages[page]

	// Check cache.
	if data, ok := z.pageCache[page]; ok {
		return data, nil
	}

	// Find the file by name in the zip.
	var target *zip.File
	for _, f := range z.reader.File {
		if f.Name == entry.Name {
			target = f
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("page file %q not found in archive", entry.Name)
	}

	rc, err := target.Open()
	if err != nil {
		return nil, fmt.Errorf("opening page %q: %w", entry.Name, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("reading page %q: %w", entry.Name, err)
	}

	z.pageCache[page] = data
	return data, nil
}

func (z *ZipArchive) Title() string {
	return z.title
}

func (z *ZipArchive) Filename() string {
	return z.filename
}

func (z *ZipArchive) Close() error {
	z.reader = nil
	z.pages = nil
	z.pageCache = nil
	return nil
}

// isImageFile returns true if the filename has a common image extension.
func isImageFile(name string) bool {
	return strings.HasSuffix(name, ".jpg") ||
		strings.HasSuffix(name, ".jpeg") ||
		strings.HasSuffix(name, ".png") ||
		strings.HasSuffix(name, ".webp") ||
		strings.HasSuffix(name, ".gif") ||
		strings.HasSuffix(name, ".bmp")
}

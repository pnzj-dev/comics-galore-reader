package cgreaderwasm

import "io"

// Archive provides a uniform interface for reading comic archives (CBZ, CBR).
type Archive interface {
	// Open initializes the archive from raw bytes and a filename.
	// The password may be empty if the archive is not encrypted.
	Open(data []byte, filename string, password string) error

	// PageCount returns the total number of pages in the archive.
	PageCount() int

	// GetPage returns the raw image bytes for a 0-based page index.
	GetPage(page int) ([]byte, error)

	// Title returns the archive title (from metadata or filename fallback).
	Title() string

	// Filename returns the original filename of the archive.
	Filename() string

	// Close releases any resources held by the archive.
	Close() error
}

// PageEntry describes a single page within an archive.
type PageEntry struct {
	Index int    // 0-based page index
	Name  string // filename within the archive (e.g. "page_001.jpg")
	Size  int64  // uncompressed size in bytes
}

// ArchiveFormat identifies the comic archive format.
type ArchiveFormat int

const (
	FormatUnknown ArchiveFormat = iota
	FormatCBZ
	FormatCBR
)

// DetectFormat detects the archive format from a filename extension.
func DetectFormat(filename string) ArchiveFormat {
	ext := ""
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			ext = filename[i+1:]
			break
		}
	}
	switch ext {
	case "cbz", "zip":
		return FormatCBZ
	case "cbr", "rar":
		return FormatCBR
	default:
		return FormatCBZ // default to CBZ / zip
	}
}

// Ensure io.Closer is referenced (used by Close methods).
var _ io.Closer = nil

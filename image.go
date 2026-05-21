package cgreaderwasm

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	_ "golang.org/x/image/webp"
)

// ImageInfo holds basic information about a decoded image.
type ImageInfo struct {
	Format      string // "jpeg", "png", "gif", "webp"
	ContentType string // MIME type for browser rendering
	Width       int
	Height      int
}

// ImageExt extracts image metadata and returns it along with a content-type
// suitable for browser rendering.
func ImageExt(data []byte) (*ImageInfo, error) {
	reader := bytes.NewReader(data)
	cfg, format, err := image.DecodeConfig(reader)
	if err != nil {
		// Try to detect format from magic bytes.
		format, ct := detectFormatFromMagic(data)
		if format == "" {
			return nil, fmt.Errorf("decoding image config: %w", err)
		}
		return &ImageInfo{
			Format:      format,
			ContentType: ct,
			Width:       0,
			Height:      0,
		}, nil
	}

	ct := contentTypeForFormat(format)
	return &ImageInfo{
		Format:      format,
		ContentType: ct,
		Width:       cfg.Width,
		Height:      cfg.Height,
	}, nil
}

// contentTypeForFormat returns the MIME content type for a given image format.
func contentTypeForFormat(format string) string {
	switch format {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// detectFormatFromMagic attempts to detect image format from magic bytes.
func detectFormatFromMagic(data []byte) (string, string) {
	if len(data) < 4 {
		return "", ""
	}
	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpeg", "image/jpeg"
	}
	// PNG: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png", "image/png"
	}
	// GIF: 47 49 46
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "gif", "image/gif"
	}
	// WebP: 52 49 46 46 ... 57 45 42 50
	if len(data) >= 12 &&
		data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return "webp", "image/webp"
	}
	return "", ""
}

// Ensure io.Reader is referenced.
var _ io.Reader = nil

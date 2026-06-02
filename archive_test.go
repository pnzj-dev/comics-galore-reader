package cgreader

import "testing"

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		filename string
		want     ArchiveFormat
	}{
		{"comic.cbz", FormatCBZ},
		{"comic.zip", FormatCBZ},
		{"comic.cbr", FormatCBR},
		{"comic.rar", FormatCBR},
		{"comic.CBZ", FormatCBZ},
		{"comic.CBR", FormatCBR},
		{"noext", FormatCBZ},
		{"", FormatCBZ},
		{"comic.cbz.backup", FormatCBZ},
	}
	for _, tt := range tests {
		got := DetectFormat(tt.filename)
		if got != tt.want {
			t.Errorf("DetectFormat(%q) = %v, want %v", tt.filename, got, tt.want)
		}
	}
}

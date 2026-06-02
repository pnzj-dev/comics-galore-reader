package cgreader

import "testing"

func TestNaturalLess(t *testing.T) {
	tests := []struct {
		a, b string
		want bool // a < b
	}{
		{"page_1.jpg", "page_2.jpg", true},
		{"page_2.jpg", "page_1.jpg", false},
		{"page_10.jpg", "page_2.jpg", false},
		{"page_2.jpg", "page_10.jpg", true},
		{"page_01.jpg", "page_1.jpg", false},
		{"page_1.jpg", "page_01.jpg", false},
		{"page_1.jpg", "page_1.jpg", false},
		{"a.jpg", "b.jpg", true},
		{"b.jpg", "a.jpg", false},
		{"1.jpg", "2.jpg", true},
		{"10.jpg", "2.jpg", false},
		{"", "a.jpg", true},
		{"a.jpg", "", false},
		{"", "", false},
		{"page_1.jpg", "page_1.png", true},
		{"page_1.png", "page_1.jpg", false},
		{"cover.jpg", "page_001.jpg", true},
		{"page_001.jpg", "cover.jpg", false},
		{"img_99.jpg", "img_100.jpg", true},
		{"img_100.jpg", "img_99.jpg", false},
	}
	for _, tt := range tests {
		got := NaturalLess(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("NaturalLess(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestSortPages(t *testing.T) {
	entries := []PageEntry{
		{Name: "page_10.jpg", Index: 0},
		{Name: "page_1.jpg", Index: 1},
		{Name: "page_2.jpg", Index: 2},
	}
	SortPages(entries)
	want := []string{"page_1.jpg", "page_2.jpg", "page_10.jpg"}
	for i, e := range entries {
		if e.Name != want[i] {
			t.Errorf("entry %d: got %q, want %q", i, e.Name, want[i])
		}
	}
}

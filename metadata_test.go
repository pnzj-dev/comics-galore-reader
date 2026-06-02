package cgreader

import "testing"

func TestParseComicInfo(t *testing.T) {
	xml := `<?xml version="1.0"?>
<ComicInfo>
  <Title>Test Comic</Title>
  <Series>Awesome Series</Series>
  <Number>42</Number>
  <Writer>Jane Doe</Writer>
  <PageCount>24</PageCount>
</ComicInfo>`

	info, err := ParseComicInfo([]byte(xml))
	if err != nil {
		t.Fatalf("ParseComicInfo failed: %v", err)
	}
	if info.Title != "Test Comic" {
		t.Errorf("expected Title='Test Comic', got %q", info.Title)
	}
	if info.Series != "Awesome Series" {
		t.Errorf("expected Series='Awesome Series', got %q", info.Series)
	}
	if info.Number != "42" {
		t.Errorf("expected Number='42', got %q", info.Number)
	}
	if info.Writer != "Jane Doe" {
		t.Errorf("expected Writer='Jane Doe', got %q", info.Writer)
	}
	if info.PageCount != 24 {
		t.Errorf("expected PageCount=24, got %d", info.PageCount)
	}
}

func TestParseComicInfoInvalidXML(t *testing.T) {
	_, err := ParseComicInfo([]byte("not xml"))
	if err == nil {
		t.Fatal("expected error for invalid XML")
	}
}

func TestParseComicInfoEmpty(t *testing.T) {
	_, err := ParseComicInfo([]byte{})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestParseComicInfoTitle(t *testing.T) {
	tests := []struct {
		xml  string
		want string
	}{
		{
			`<?xml version="1.0"?><ComicInfo><Title>Single</Title></ComicInfo>`,
			"Single",
		},
		{
			`<?xml version="1.0"?><ComicInfo><Series>Run</Series><Number>5</Number></ComicInfo>`,
			"Run #5",
		},
		{
			`<?xml version="1.0"?><ComicInfo><Series>Run</Series></ComicInfo>`,
			"Run",
		},
		{
			`<?xml version="1.0"?><ComicInfo></ComicInfo>`,
			"",
		},
		{
			`invalid`,
			"",
		},
		{
			``,
			"",
		},
	}
	for _, tt := range tests {
		got := ParseComicInfoTitle([]byte(tt.xml))
		if got != tt.want {
			t.Errorf("ParseComicInfoTitle(%q) = %q, want %q", tt.xml, got, tt.want)
		}
	}
}

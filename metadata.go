package cgreaderwasm

import (
	"encoding/xml"
)

// ComicInfo represents the ComicInfo.xml metadata commonly found in comic archives.
type ComicInfo struct {
	XMLName xml.Name `xml:"ComicInfo"`
	Title   string   `xml:"Title"`
	Series  string   `xml:"Series"`
	Number  string   `xml:"Number"`
	Volume  int      `xml:"Volume"`
	Summary string   `xml:"Summary"`
	Writer  string   `xml:"Writer"`
	Penciller string `xml:"Penciller"`
	Inker   string   `xml:"Inker"`
	Colorist string  `xml:"Colorist"`
	Letterer string  `xml:"Letterer"`
	CoverArtist string `xml:"CoverArtist"`
	Publisher string `xml:"Publisher"`
	Genre    string   `xml:"Genre"`
	PageCount int    `xml:"PageCount"`
	Year     int     `xml:"Year"`
	Month    int     `xml:"Month"`
	Day      int     `xml:"Day"`
}

// ParseComicInfo parses ComicInfo.xml data and returns the structured metadata.
func ParseComicInfo(data []byte) (*ComicInfo, error) {
	var info ComicInfo
	if err := xml.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// ParseComicInfoTitle extracts just the title from ComicInfo.xml data.
// Returns empty string if parsing fails or title is missing.
func ParseComicInfoTitle(data []byte) string {
	info, err := ParseComicInfo(data)
	if err != nil || info == nil {
		return ""
	}
	// Prefer Series + Number combo, fallback to Title.
	if info.Series != "" {
		if info.Number != "" {
			return info.Series + " #" + info.Number
		}
		return info.Series
	}
	return info.Title
}

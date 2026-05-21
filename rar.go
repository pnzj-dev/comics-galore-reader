package cgreaderwasm

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/nwaples/rardecode/v2"
)

// RarArchive implements Archive for CBR files (RAR format).
type RarArchive struct {
	filename  string
	title     string
	pages     []PageEntry
	data      []byte
	password  string
	pageCache map[int][]byte
}

func (r *RarArchive) Open(data []byte, filename string, password string) error {
	r.filename = filename
	r.pages = nil
	r.pageCache = make(map[int][]byte)
	r.data = data
	r.password = password

	var opts []rardecode.Option
	if password != "" {
		opts = append(opts, rardecode.Password(password))
	}

	reader, err := rardecode.NewReader(bytes.NewReader(data), opts...)
	if err != nil {
		return fmt.Errorf("opening rar: %w", err)
	}
	

	var pageEntries []PageEntry
	index := 0
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading rar entry: %w", err)
		}

		if header.IsDir {
			continue
		}

		lower := strings.ToLower(header.Name)
		if isImageFile(lower) {
			pageEntries = append(pageEntries, PageEntry{
				Index: index,
				Name:  header.Name,
				Size:  header.UnPackedSize,
			})
			index++
		} else if lower == "comicinfo.xml" {
			xmlData, err := io.ReadAll(reader)
			if err == nil {
				r.title = ParseComicInfoTitle(xmlData)
			}
		}
	}

	SortPages(pageEntries)
	for i := range pageEntries {
		pageEntries[i].Index = i
	}
	r.pages = pageEntries

	return nil
}

func (r *RarArchive) PageCount() int {
	return len(r.pages)
}

func (r *RarArchive) GetPage(page int) ([]byte, error) {
	if page < 0 || page >= len(r.pages) {
		return nil, fmt.Errorf("page %d out of range [0, %d)", page, len(r.pages))
	}

	if data, ok := r.pageCache[page]; ok {
		return data, nil
	}

	entry := r.pages[page]

	var opts []rardecode.Option
	if r.password != "" {
		opts = append(opts, rardecode.Password(r.password))
	}

	reader, err := rardecode.NewReader(bytes.NewReader(r.data), opts...)
	if err != nil {
		return nil, fmt.Errorf("re-opening rar: %w", err)
	}
	

	// Seek to the target entry.
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("page %q not found in rar", entry.Name)
		}
		if err != nil {
			return nil, fmt.Errorf("seeking rar: %w", err)
		}
		if header.Name == entry.Name {
			data, err := io.ReadAll(reader)
			if err != nil {
				return nil, fmt.Errorf("reading page %q: %w", entry.Name, err)
			}
			r.pageCache[page] = data
			return data, nil
		}
	}
}

func (r *RarArchive) Title() string {
	return r.title
}

func (r *RarArchive) Filename() string {
	return r.filename
}

func (r *RarArchive) Close() error {
	r.data = nil
	r.pages = nil
	r.pageCache = nil
	return nil
}

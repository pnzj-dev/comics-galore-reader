package cgreaderwasm

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultPasswordURL is the default endpoint for fetching archive passwords.
const DefaultPasswordURL = "/api/comic-password"

// FetchPassword fetches a password string from the given URL.
// It expects the response body to contain the password as plain text (trimmed).
func FetchPassword(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("password URL is empty")
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http get %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("password fetch returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096)) // 4KB max
	if err != nil {
		return "", fmt.Errorf("reading password response: %w", err)
	}

	password := strings.TrimSpace(string(body))
	if password == "" {
		return "", fmt.Errorf("empty password received from %q", url)
	}

	return password, nil
}

// SetPasswordURL is a convenience method exposed to JavaScript for runtime override.
// It delegates to ComicReader.SetPasswordURL.
func (r *ComicReader) SetPassword(url string) {
	r.SetPasswordURL(url)
}

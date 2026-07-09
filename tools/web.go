package tools

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"time"
)

type FetchURLTool struct{}

func (t *FetchURLTool) Name() string {
	return "fetch_url"
}

func (t *FetchURLTool) Description() string {
	return `TOOL: fetch_url (To read text content from a specific website URL)
        {
            "thinking": "I need to read the contents of this URL to gather information.",
            "action": "fetch_url",
            "data": "https://example.com",
            "content": ""
        }`
}

func (t *FetchURLTool) Execute(data, content string) string {
	// Explicit timeout prevents goroutine leaks if a website hangs indefinitely.
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", data, nil)
	if err != nil {
		return "ERROR: İnvalid URL - " + err.Error()
	}

	// Spoof User-Agent to bypass simple anti-bot mechanisms (like default 403 blocks for Go-http-client).
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "ERROR: Failed to connect to the site - " + err.Error()
	}
	defer resp.Body.Close()

	// io.LimitReader acts as a strict safety valve. It caps the reading at 50 KB,
	// protecting the Heap memory from OOM crashes if the agent tries to download huge files (e.g., ISOs).
	limitedReader := io.LimitReader(resp.Body, 50*1024)
	bodyBytes, _ := io.ReadAll(limitedReader)

	// Regex compilation is CPU-heavy. However, since bodyBytes is strictly bounded (50KB max),
	// this O(N) operation executes in mere milliseconds, preventing CPU spikes.
	re := regexp.MustCompile(`<[^>]*>`)
	cleanText := re.ReplaceAllString(string(bodyBytes), " ")

	if len(cleanText) == 0 {
		return "ERROR: Site is empty or no text found."
	}

	return "SYSTEM DATA (Website Content):\n" + cleanText
}

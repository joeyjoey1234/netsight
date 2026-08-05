package scan

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func GrabHTTPTitle(ctx context.Context, target string) string {
	for _, port := range []int{80, 443, 8080, 8443} {
		title := grabTitle(ctx, target, port)
		if title != "" {
			return title
		}
	}
	return ""
}

func grabTitle(ctx context.Context, target string, port int) string {
	url := fmt.Sprintf("http://%s:%d/", target, port)
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{Timeout: 2 * time.Second}).DialContext,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ""
	}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body := make([]byte, 16384)
	n, _ := io.ReadFull(resp.Body, body)
	content := string(body[:n])

	title := extractTitle(content)
	if title == "" {
		return ""
	}
	return fmt.Sprintf("[%s:%d] %s", target, port, title)
}

func extractTitle(html string) string {
	lower := strings.ToLower(html)
	startTag := "<title"
	endTag := "</title>"

	start := strings.Index(lower, startTag)
	if start == -1 {
		return ""
	}

	closeBracket := strings.Index(lower[start:], ">")
	if closeBracket == -1 {
		return ""
	}
	contentStart := start + closeBracket + 1

	end := strings.Index(lower[contentStart:], endTag)
	if end == -1 {
		return ""
	}

	title := html[contentStart : contentStart+end]
	return strings.TrimSpace(title)
}

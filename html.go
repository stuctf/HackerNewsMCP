package main

import (
	"html"
	"regexp"
	"strings"
)

var (
	reLink = regexp.MustCompile(`<a\s[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)
	reTag  = regexp.MustCompile(`<[^>]+>`)
)

// StripHTML converts HN comment HTML to plain text with markdown links.
func StripHTML(s string) string {
	// Convert <p> tags to double newlines.
	s = strings.ReplaceAll(s, "<p>", "\n\n")

	// Convert <a href="...">text</a> to [text](url).
	s = reLink.ReplaceAllString(s, "[$2]($1)")

	// Remove remaining HTML tags.
	s = reTag.ReplaceAllString(s, "")

	// Decode HTML entities.
	s = html.UnescapeString(s)

	return strings.TrimSpace(s)
}

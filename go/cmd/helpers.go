package cmd

import (
	"fmt"
	"regexp"
	"strings"
)

var tweetURLRe = regexp.MustCompile(`(?:twitter\.com|x\.com)/\w+/status/(\d+)`)
var numericRe = regexp.MustCompile(`^\d+$`)

// parseTweetID extracts a tweet ID from a URL or raw numeric string.
func parseTweetID(input string) (string, error) {
	if m := tweetURLRe.FindStringSubmatch(input); m != nil {
		return m[1], nil
	}
	s := strings.TrimSpace(input)
	if numericRe.MatchString(s) {
		return s, nil
	}
	return "", fmt.Errorf("invalid tweet ID or URL: %s", input)
}

// stripAt removes a leading @ from a username.
func stripAt(username string) string {
	return strings.TrimPrefix(username, "@")
}

// formatNumber formats an integer with comma separators.
func formatNumber(n int) string {
	if n < 0 {
		return "-" + formatNumber(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

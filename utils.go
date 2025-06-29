package main

import (
	"net/url"
	"strings"
)

// isValidURL checks if a string is a valid URL
func isValidURL(input string) bool {
	// Add scheme if missing
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		input = "https://" + input
	}

	_, err := url.ParseRequestURI(input)
	return err == nil
}

// formatURL ensures a URL has the proper scheme
func formatURL(input string) string {
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return "https://" + input
	}
	return input
}

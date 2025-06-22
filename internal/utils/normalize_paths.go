package utils

import (
	"strings"
)

func NormalizePath(path string) string {
	path = strings.ToLower(path)
	path = strings.TrimRight(path, "/")

	if path == "" {
		path = "/"
	}

	var b strings.Builder
	last := rune(0)
	for _, r := range path {
		if r != '/' || last != '/' {
			b.WriteRune(r)
		}
		last = r
	}
	return b.String()
}

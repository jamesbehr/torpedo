package util

import (
	"os"
	"path/filepath"
)

func ExpandPath(p string) string {
	if p == "" {
		return ""
	}

	home := os.Getenv("HOME")

	if filepath.IsAbs(home) {
		rel, err := filepath.Rel("~", p)
		if err == nil && filepath.IsLocal(rel) {
			return filepath.Join(home, rel)
		}
	}

	return p
}

func UnexpandPath(p string) string {
	if p == "" {
		return ""
	}

	home := os.Getenv("HOME")

	if filepath.IsAbs(home) {
		rel, err := filepath.Rel(home, p)
		if err == nil && filepath.IsLocal(rel) {
			return filepath.Join("~", rel)
		}
	}

	return p
}

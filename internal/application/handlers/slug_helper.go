// internal/application/handlers/slug_helper.go
package handlers

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// GenerateSlug genera un slug URL-friendly desde un nombre
func GenerateSlug(name string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, name)
	result = strings.ToLower(result)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	result = reg.ReplaceAllString(result, "-")
	result = strings.Trim(result, "-")
	return result
}

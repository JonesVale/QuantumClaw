package common

import "strings"

// NormalizeModelName normalizes a model display name to the format used
// as keys in the ModelRatio map: lowercase, spaces to hyphens, strip
// Unicode whitespace and underscores.
func NormalizeModelName(name string) string {
	// Replace Unicode non-breaking spaces and zero-width spaces
	name = strings.ReplaceAll(name, "\u00a0", " ")
	name = strings.ReplaceAll(name, "\u200b", "")
	// Replace underscores with hyphens
	name = strings.ReplaceAll(name, "_", "-")
	// Replace tabs with spaces (will become hyphens below)
	name = strings.ReplaceAll(name, "\t", " ")
	// Lowercase
	name = strings.ToLower(name)
	// Spaces to hyphens
	name = strings.ReplaceAll(name, " ", "-")
	return name
}

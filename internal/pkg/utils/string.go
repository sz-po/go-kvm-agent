package utils

import "regexp"

var kebabCaseRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// IsKebabCase checks if a string follows kebab-case naming convention.
func IsKebabCase(s string) bool {
	return kebabCaseRegex.MatchString(s)
}

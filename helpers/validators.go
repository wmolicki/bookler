package helpers

import "strings"

// CommaSplit will split string s on commas, adding some additional cleaning
// such as remove trailing spaces
func CommaSplit(s string) []string {
	split := strings.Split(s, ",")
	for i, ss := range split {
		split[i] = strings.TrimSpace(ss)
	}
	return split
}

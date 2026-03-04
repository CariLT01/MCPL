package utils

import (
	"path/filepath"
)

func SliceToMap(elements []string) map[string]struct{} {
	// Pre-allocate map size to avoid re-allocations during growth
	m := make(map[string]struct{}, len(elements))

	for _, s := range elements {
		// fmt.Println("add: " + s)
		m[filepath.Clean(s)] = struct{}{}
	}
	return m
}

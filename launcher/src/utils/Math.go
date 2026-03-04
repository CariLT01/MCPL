package utils

import "fmt"

func FormatSeconds(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

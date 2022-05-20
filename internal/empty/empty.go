package empty

import "strings"

func String(s string) bool {
	return strings.TrimSpace(s) == ""
}

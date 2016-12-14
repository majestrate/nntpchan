package util

import "strings"

func IsSage(str string) bool {
	str = strings.ToLower(str)
	return str == "sage" || strings.HasPrefix(str, "sage ")
}

package v1

import "strings"

func ToLowerInPlace(ss []string) {
	for i, v := range ss {
		ss[i] = strings.ToLower(v)
	}
}

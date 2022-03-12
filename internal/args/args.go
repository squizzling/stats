package args

import (
	"strings"
)

func Flatten(ss []string) []string {
	var out []string
	for _, s := range ss {
		out = append(out, strings.Split(s, ",")...)
	}
	return out
}



package glob

import (
	"strings"
)

type globMatcher struct {
	parts  []string
	invert bool
	prefix string
	suffix string
}

func (gm *globMatcher) Match(input string) bool {
	if !strings.HasPrefix(input, gm.prefix) {
		return gm.invert
	}
	if !strings.HasSuffix(input, gm.suffix) {
		return gm.invert
	}

	startIdx := 0
	for _, part := range gm.parts {
		idx := strings.Index(input[startIdx:], part)
		if idx == -1 {
			return gm.invert
		}
		startIdx += idx
	}
	return !gm.invert
}

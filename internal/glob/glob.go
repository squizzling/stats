package glob

import (
	"strings"
)

type Matcher interface {
	Match(input string) bool
}

func NewMatcher(pattern string) Matcher {
	if pattern == "" {
		return &globEmpty{}
	}
	if pattern == "!" {
		return &globEmptyInvert{}
	}

	invert := pattern[0] == '!'
	if invert {
		pattern = pattern[1:]
	}

	// TODO: Deal with foo**bar -> foo*bar, replace doesn't collapse multiple
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		if invert {
			return &globExactInvert{
				part: parts[0],
			}
		} else {
			return &globExact{
				part: parts[0],
			}
		}
	}

	var prefix, suffix string
	if !strings.HasPrefix(pattern, "*") {
		prefix = parts[0]
		parts = parts[1:]
	}
	if !strings.HasSuffix(pattern, "*") {
		suffix = parts[len(parts)-1]
	}
	return &globMatcher{
		parts:  parts,
		invert: invert,
		prefix: prefix,
		suffix: suffix,
	}
}

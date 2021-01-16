package glob

type globEmptyInvert struct{}

func (ge *globEmptyInvert) Match(input string) bool {
	return input != ""
}

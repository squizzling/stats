package glob

type globExactInvert struct {
	part string
}

func (gne *globExactInvert) Match(input string) bool {
	return input != gne.part
}

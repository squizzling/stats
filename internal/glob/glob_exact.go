package glob

type globExact struct {
	part string
}

func (ge *globExact) Match(input string) bool {
	return input == ge.part
}

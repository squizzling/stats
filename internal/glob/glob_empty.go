package glob

type globEmpty struct{}

func (ge *globEmpty) Match(input string) bool {
	return input == ""
}

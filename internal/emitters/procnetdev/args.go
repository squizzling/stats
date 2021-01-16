package procnetdev

type ProcNetDevOpts struct {
	IncludeInterface []string `long:"include-interface"`
	ExcludeInterface []string `long:"exclude-interface"`
	IncludeContainer []string `long:"include-container"`
	ExcludeContainer []string `long:"exclude-container"`
}

func (opts *ProcNetDevOpts) Validate() []string {
	return nil
}

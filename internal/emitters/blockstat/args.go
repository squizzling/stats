package blockstat

type BlockStatOpts struct {
	IncludeDevice []string `long:"include-device"`
	ExcludeDevice []string `long:"exclude-device"`
}

func (opts *BlockStatOpts) Validate() []string {
	return nil
}

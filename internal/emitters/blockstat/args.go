package blockstat

import (
	"github.com/squizzling/stats/internal/args"
)

type BlockStatOpts struct {
	IncludeDevice []string `long:"blockstat.include-device"`
	ExcludeDevice []string `long:"blockstat.exclude-device"`
}

func (opts *BlockStatOpts) Validate() []string {
	opts.IncludeDevice = args.Flatten(opts.IncludeDevice)
	opts.ExcludeDevice = args.Flatten(opts.ExcludeDevice)
	return nil
}

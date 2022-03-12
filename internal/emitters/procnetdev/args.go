package procnetdev

import (
	"github.com/squizzling/stats/internal/args"
)

type ProcNetDevOpts struct {
	IncludeInterface []string `long:"procnetdev.include-interface"`
	ExcludeInterface []string `long:"procnetdev.exclude-interface"`
	IncludeContainer []string `long:"procnetdev.include-container"`
	ExcludeContainer []string `long:"procnetdev.exclude-container"`
}

func (opts *ProcNetDevOpts) Validate() []string {

	opts.IncludeInterface = args.Flatten(opts.IncludeInterface)
	opts.ExcludeInterface = args.Flatten(opts.ExcludeInterface)
	opts.IncludeContainer = args.Flatten(opts.IncludeContainer)
	opts.ExcludeContainer = args.Flatten(opts.ExcludeContainer)
	return nil
}

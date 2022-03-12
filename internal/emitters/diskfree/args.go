package diskfree

import (
	"github.com/squizzling/stats/internal/args"
)

type DiskFreeOpts struct {
	IncludeMountPoint []string `long:"diskfree.include-mount-point"`
	ExcludeMountPoint []string `long:"diskfree.exclude-mount-point"`
	IncludeFsType     []string `long:"diskfree.include-fs-type"`
	ExcludeFsType     []string `long:"diskfree.exclude-fs-type"`
}

func (opts *DiskFreeOpts) Validate() []string {
	opts.IncludeFsType = args.Flatten(opts.IncludeFsType)
	opts.ExcludeFsType = args.Flatten(opts.ExcludeFsType)
	opts.IncludeMountPoint = args.Flatten(opts.IncludeMountPoint)
	opts.ExcludeMountPoint = args.Flatten(opts.ExcludeMountPoint)
	if len(opts.IncludeFsType) == 0 && len(opts.ExcludeFsType) == 0 {
		opts.ExcludeFsType = []string{
			"autofs",
			"aufs",
			"bpf",
			"binfmt_misc",
			"cgroup",
			"cgroup2",
			"configfs",
			"debugfs",
			"devpts",
			"devtmpfs",
			"efivarfs",
			"fusectl",
			"hugetlbfs",
			"mqueue",
			"nfsd",
			"overlay",
			"proc",
			"pstore",
			"rpc_pipefs",
			"securityfs",
			"sysfs",
			"tmpfs",
			"tracefs",
		}
	}
	return nil
}

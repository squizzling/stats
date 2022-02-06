package diskfree

type DiskFreeOpts struct {
	IncludeMountPoint []string `long:"diskfree.include-mount-point"`
	ExcludeMountPoint []string `long:"diskfree.exclude-mount-point"`
	IncludeFsType     []string `long:"diskfree.include-fs-type"`
	ExcludeFsType     []string `long:"diskfree.exclude-fs-type"`
}

func (opts *DiskFreeOpts) Validate() []string {
	return nil
}

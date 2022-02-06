package bucketstat

import (
	"net/url"
	"strings"
)

type BucketStatOpts struct {
	Frequency int      `long:"bucketstat.frequency" default:"1" description:"how many emit intervals to run, must be positive"`
	Profile   *string  `long:"bucketstat.profile"               description:"aws profile to load"                             `
	Prefix    []string `long:"bucketstat.prefix"                description:"prefix to match, in form s3://bucket[/prefix]"   `

	prefixes []bucketAndPrefix
}

func (opts *BucketStatOpts) Validate() []string {
	var errs []string
	if opts.Frequency <= 0 {
		errs = append(errs, "bucketstat.frequency must be positive")
	}
	for _, prefix := range opts.Prefix {
		u, err := url.Parse(prefix)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		if u.Scheme != "s3" {
			errs = append(errs, "bucketstat.prefix must start with s3://")
			continue
		}
		if u.Host == "" {
			errs = append(errs, "bucketstat.prefix must contain a bucket")
			continue
		}

		opts.prefixes = append(opts.prefixes, bucketAndPrefix{
			bucket: u.Host,
			prefix: strings.TrimLeft(u.Path, "/"),

		})
	}
	return errs
}

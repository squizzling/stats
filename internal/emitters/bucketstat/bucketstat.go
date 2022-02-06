package bucketstat

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/emitters/bucketstat/paginator"
	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type bucketAndPrefix struct {
	bucket string
	prefix string
}

type BucketStatEmitter struct {
	logger    *zap.Logger
	statsPool statser.Pool

	prefix []bucketAndPrefix

	frequency int
	next      int

	s3client *s3.Client
}

func NewEmitter(logger *zap.Logger, statsPool statser.Pool, opt emitter.OptProvider) emitter.Emitter {
	opts := opt.Get("bucketstat").(*BucketStatOpts)

	var awsOpts []func(*config.LoadOptions) error
	if opts.Profile != nil {
		awsOpts = append(awsOpts, config.WithSharedConfigProfile(*opts.Profile))
	}
	cfg, err := config.LoadDefaultConfig(context.Background(), awsOpts...)
	if err != nil {
		panic(err)
	}

	s3client := s3.NewFromConfig(cfg)

	return &BucketStatEmitter{
		prefix:    opts.prefixes,
		frequency: opts.Frequency,
		logger:    logger,
		statsPool: statsPool,
		s3client:  s3client,
	}
}

func (bse *BucketStatEmitter) Emit() {
	bse.next++
	if bse.next < bse.frequency {
		return
	}
	bse.next = 0
	calls := 0
	for _, p := range bse.prefix {
		calls += bse.EmitPrefix(p.bucket, p.prefix)
	}
	bse.statsPool.Global().Count("bucketstat.calls", calls)
}

func (bse *BucketStatEmitter) EmitPrefix(bucket, prefix string) int {
	pager := paginator.NewListObjectVersionsPaginator(bse.s3client, &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	objSizes := make(objSizeMap)
	activeObjectCount := 0
	deadObjectCount := 0
	var latest time.Time

	for pager.HasMorePages() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			bse.logger.Error("NextPage", zap.Error(err))
			return pager.Calls()
		}
		for _, version := range page.Versions {
			objSizes.get(*version.Key).size += version.Size
			if version.IsLatest {
				activeObjectCount++
			}
			if latest.Before(*version.LastModified) {
				latest = *version.LastModified
			}
		}

		for _, deleted := range page.DeleteMarkers {
			if deleted.IsLatest {
				objSizes.get(*deleted.Key).deleted = true
				deadObjectCount++
			}
			if latest.Before(*deleted.LastModified) {
				latest = *deleted.LastModified
			}
		}
	}

	deletedBytes := int64(0)
	activeBytes := int64(0)
	for _, o := range objSizes {
		if o.deleted {
			deletedBytes += o.size
		} else {
			activeBytes += o.size
		}
	}

	bse.statsPool.Global("bucket", bucket, "prefix", "/"+prefix, "state", "active").Gauge("bucketstat.bytes", activeBytes)
	bse.statsPool.Global("bucket", bucket, "prefix", "/"+prefix, "state", "active").Gauge("bucketstat.objects", activeObjectCount)
	bse.statsPool.Global("bucket", bucket, "prefix", "/"+prefix, "state", "deleted").Gauge("bucketstat.bytes", deletedBytes)
	bse.statsPool.Global("bucket", bucket, "prefix", "/"+prefix, "state", "deleted").Gauge("bucketstat.objects", deadObjectCount)
	bse.statsPool.Global("bucket", bucket, "prefix", "/"+prefix).Gauge("bucketstat.latest", latest.UnixNano()/1000000)
	fmt.Printf("%d %d %d %d %d\n", activeBytes, activeObjectCount, deletedBytes, deadObjectCount, latest.UnixNano()/1000000)

	return pager.Calls()
}

func init() {
	sources.Sources["bucketstat"] = NewEmitter
}

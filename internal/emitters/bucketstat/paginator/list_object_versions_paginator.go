package paginator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ListObjectVersionsPaginator struct {
	client *s3.Client
	lovi   *s3.ListObjectVersionsInput

	calls int

	isComplete bool
}

func NewListObjectVersionsPaginator(client *s3.Client, params *s3.ListObjectVersionsInput) *ListObjectVersionsPaginator {
	return &ListObjectVersionsPaginator{
		client: client,
		lovi:   params,
	}
}

func (lovp *ListObjectVersionsPaginator) HasMorePages() bool {
	return !lovp.isComplete
}

func (lovp *ListObjectVersionsPaginator) NextPage(ctx context.Context) (*s3.ListObjectVersionsOutput, error) {
	if !lovp.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	lovo, err := lovp.client.ListObjectVersions(ctx, lovp.lovi)
	lovp.calls++
	if err != nil {
		return nil, err
	}
	if !lovo.IsTruncated {
		lovp.isComplete = true
	} else {
		lovp.lovi.KeyMarker = lovo.NextKeyMarker
		lovp.lovi.VersionIdMarker = lovo.NextVersionIdMarker
	}
	return lovo, nil
}

func (lovp *ListObjectVersionsPaginator) Calls() int {
	return lovp.calls
}
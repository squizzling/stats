package istats

import (
	"github.com/squizzling/stats/pkg/statser"
)

var _ = statser.Pool(&fakePool{})

type fakePool struct {
	hostName string
}

func NewFakePool(hostName string) statser.Pool {
	return &fakePool{
		hostName: hostName,
	}
}

func (f *fakePool) Host(tags ...string) statser.Statser {
	return &fakeStatser{
		tags: append([]string{"host", f.hostName}, tags...),
	}
}

func (f *fakePool) Global(tags ...string) statser.Statser {
	return &fakeStatser{
		tags: tags,
	}
}

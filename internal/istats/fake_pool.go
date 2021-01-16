package istats

import (
	"github.com/squizzling/stats/pkg/statser"
)

var _ = statser.Pool(&fakePool{})

type fakePool struct{}

func NewFakePool() statser.Pool {
	return &fakePool{}
}

func (f *fakePool) Get(tags ...string) statser.Statser {
	return &fakeStatser{
		tags: tags,
	}
}

func (f *fakePool) GetFake(tags ...string) statser.Statser {
	return &fakeStatser{
		tags: tags,
	}
}

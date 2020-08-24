package istats

import (
	"strings"

	"github.com/alexcesaro/statsd"

	"github.com/squizzling/stats/pkg/statser"
)

var _ = statser.Pool(&Pool{})
var _ = statser.Statser(&statsd.Client{})

type Pool struct {
	base    *statsd.Client
	clients map[string]*statsd.Client
}

func NewPool(c *statsd.Client) *Pool {
	return &Pool{
		base: c,
		clients: map[string]*statsd.Client{
			"": c,
		},
	}
}

func (p *Pool) Get(tags ...string) statser.Statser {
	s := strings.Join(tags, "||")
	if c, ok := p.clients[s]; ok {
		return c
	}
	c := p.base.Clone(statsd.Tags(tags...))
	p.clients[s] = c
	return c
}

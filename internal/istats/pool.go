package istats

import (
	"strings"

	"github.com/alexcesaro/statsd"

	"github.com/squizzling/stats/pkg/statser"
)

var _ = statser.Pool(&Pool{})
var _ = statser.Statser(&statsd.Client{})

type Pool struct {
	hostName string
	base     *statsd.Client
	clients  map[string]*statsd.Client
}

func NewPool(hostName string, c *statsd.Client) *Pool {
	return &Pool{
		hostName: hostName,
		base:     c,
		clients:  map[string]*statsd.Client{},
	}
}

func (p *Pool) Host(tags ...string) statser.Statser {
	return p.Global(append([]string{"host", p.hostName}, tags...)...)
}

func (p *Pool) Global(tags ...string) statser.Statser {
	s := strings.Join(tags, "||")
	if c, ok := p.clients[s]; ok {
		return c
	}
	c := p.base.Clone(statsd.Tags(tags...))
	p.clients[s] = c
	return c
}

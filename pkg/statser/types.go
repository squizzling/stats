package statser

type Pool interface {
	Get(tags ...string) Statser
}

type Statser interface {
	Gauge(metricName string, value interface{})
}
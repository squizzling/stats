package statser

type Pool interface {
	Host(tags ...string) Statser
	Global(tags ...string) Statser
}

type Statser interface {
	Gauge(metricName string, value interface{})
}

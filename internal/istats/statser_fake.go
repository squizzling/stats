package istats

import (
	"fmt"
	"strings"
)

type fakeStatser struct {
	tags []string
}

func (fs *fakeStatser) Gauge(metricName string, metricValue interface{}) {
	sb := strings.Builder{}
	for i := 0; i < len(fs.tags); i += 2 {
		if i != 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(fs.tags[i+0])
		sb.WriteByte('=')
		sb.WriteString(fs.tags[i+1])
	}

	fmt.Printf("Gauge: %s{%s}=%v\n", metricName, sb.String(), metricValue)
}

func (fs *fakeStatser) Count(metricName string, metricValue interface{}) {
	sb := strings.Builder{}
	for i := 0; i < len(fs.tags); i += 2 {
		if i != 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(fs.tags[i+0])
		sb.WriteByte('=')
		sb.WriteString(fs.tags[i+1])
	}

	fmt.Printf("Count: %s{%s}=%v\n", metricName, sb.String(), metricValue)
}

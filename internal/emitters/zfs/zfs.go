package zfs

import (
	"fmt"
	"io/ioutil"

	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/emitters/zfs/kstat"
	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type ZFSEmitter struct {
	logger    *zap.Logger
	statsPool statser.Pool
}

func NewEmitter(logger *zap.Logger, statsPools statser.Pool) emitter.Emitter {
	return &ZFSEmitter{
		logger:    logger,
		statsPool: statsPools,
	}
}

func (e *ZFSEmitter) statPoolIo(poolName string) *kstat.Kstat {
	/*
		typedef struct kstat_io {           // *** More details are in the struct definition
			u_longlong_t    nread;          // number of bytes read
			u_longlong_t    nwritten;       // number of bytes written
			uint_t          reads;          // number of read operations
			uint_t          writes;         // number of write operations
			hrtime_t        wtime;          // cumulative wait (pre-service) time
			hrtime_t        wlentime;       // cumulative wait len*time product
			hrtime_t        wlastupdate;    // last time wait queue changed
			hrtime_t        rtime;          // cumulative run (service) time
			hrtime_t        rlentime;       // cumulative run length*time product
			hrtime_t        rlastupdate;    // last time run queue changed
			uint_t          wcnt;           // count of elements in wait state
			uint_t          rcnt;           // count of elements in run state
		} kstat_io_t;
	*/
	return kstat.LoadKstat(fmt.Sprintf("/proc/spl/kstat/zfs/%s/io", poolName), e.logger)
}

func (e *ZFSEmitter) statArc() *kstat.Kstat {
	return kstat.LoadKstat("/proc/spl/kstat/zfs/arcstats", e.logger)
}

func (e *ZFSEmitter) statPools() map[string]*kstat.Kstat {
	results := make(map[string]*kstat.Kstat)

	poolDirs, _ := ioutil.ReadDir("/proc/spl/kstat/zfs/")
	for _, stat := range poolDirs {
		if stat.IsDir() {
			poolName := stat.Name()
			kst := e.statPoolIo(poolName)
			if kst != nil {
				results[poolName] = kst
			}
		}
	}
	return results
}

func (e *ZFSEmitter) Emit() {
	kst := e.statArc()
	for k, v := range kst.UValues {
		metricName := "zfs.arc." + k
		e.statsPool.Get().Gauge(metricName, v)
	}

	for poolName, kst := range e.statPools() {
		client := e.statsPool.Get("pool", poolName)
		for k, v := range kst.UValues {
			metricName := "zfs.io." + k
			client.Gauge(metricName, v)
		}
	}
}

func init() {
	sources.Sources["zfs"] = NewEmitter
}

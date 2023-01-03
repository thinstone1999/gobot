package metric

import (
	"hash/crc32"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/Gonewithmyself/gobot/pkg/logger"
)

type IMetric interface {
	GetGamer() string      // 指标所属的玩家
	GetMsgId() interface{} // 指标所属的消息
}

type IMetricHandler interface {
	ProcessMetric(metric IMetric, rec *Recorder) // 处理指标, rec对应全局信息
	Report()                                     // 输出报告
	TransMsgId(interface{}) string               // 转换msgId为可读
	GetReportFile() string                       // 报告文件名
}

func Start(handler IMetricHandler) *Manager {
	Mgr = &Manager{
		handler: handler,
		done:    make(chan struct{}),
	}

	Mgr.Start()
	return Mgr
}

// 记录指标
// 注意需要过滤一些无关紧要的消息例如心跳 避免负载过高
func Record(metric IMetric) {
	atomic.AddInt64(&Mgr.pending, 1)
	sum := crc32.Checksum([]byte(metric.GetGamer()), crc32.IEEETable)
	idx := sum % uint32(len(Mgr.workers))
	Mgr.workers[idx].submit(metric)
}

func Stop() {
	Mgr.Once.Do(func() {
		for _, worker := range Mgr.workers {
			close(worker.donec)
		}
		logger.Debug("notify worker exit")
		Mgr.wg.Wait()
		logger.Info("metric exit", "pending", Mgr.pending)
		Mgr.handler.Report()
	})
}

type Manager struct {
	workers []*Worker
	done    chan struct{}
	pending int64 // 待处理的指标
	sync.Once
	wg      sync.WaitGroup
	handler IMetricHandler
}

var Mgr *Manager

func (mgr *Manager) Start() {
	n := runtime.NumCPU()
	for i := 0; i < n; i++ {
		id := int32(i + 1)
		w := NewWorker(id)
		mgr.workers = append(mgr.workers, w)
		mgr.wg.Add(1)
		go func() {
			w.run()
			mgr.wg.Done()
		}()
	}

	logger.Info("metric start", "worker", n)
}

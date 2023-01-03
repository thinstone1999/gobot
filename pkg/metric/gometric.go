package metric

import (
	"fmt"
	"sync"

	"github.com/rcrowley/go-metrics"
)

// 各消息情况
var RecorderMap sync.Map // map[msgid]*Recorder

type Recorder struct {
	UpCounter   metrics.Counter // 请求数量
	DownCounter metrics.Counter // 回包数量
	RTTRecorder metrics.Timer   // 往返时延
}

func NewRecorder(cmd interface{}) *Recorder {
	str := fmt.Sprintf("%v", cmd)
	UpCounter := metrics.NewCounter()
	DownCounter := metrics.NewCounter()
	upval := metrics.GetOrRegister("Up_"+str, UpCounter)
	downVal := metrics.GetOrRegister("Down_"+str, DownCounter)

	tmpRtt := metrics.NewTimer()
	rtt := metrics.GetOrRegister("RTT_"+str, tmpRtt)
	return &Recorder{
		RTTRecorder: rtt.(metrics.Timer),
		UpCounter:   upval.(metrics.Counter),
		DownCounter: downVal.(metrics.Counter),
	}
}

type ReportItem struct {
	UpCount     int64  // 发包数量
	DownCount   int64  // 收包数量
	RspRate     string // 响应率
	Min         string // 最小时延 毫秒
	Max         string // 最大时延
	Avg         string // 平均时延
	Fifty       string // 时延50%分位
	SeventyFive string // 时延75%分位
	Ninety      string // 时延90%分位
}

func Status() map[string]*ReportItem {
	var mp = make(map[string]*ReportItem)
	RecorderMap.Range(func(key, value interface{}) bool {
		rec := value.(*Recorder)
		item := &ReportItem{
			UpCount:     rec.UpCounter.Count(),
			DownCount:   rec.DownCounter.Count(),
			Min:         fmt.Sprintf("%.2f", float64(rec.RTTRecorder.Min())/1e6),
			Max:         fmt.Sprintf("%.2f", float64(rec.RTTRecorder.Max())/1e6),
			Avg:         fmt.Sprintf("%.2f", float64(rec.RTTRecorder.Mean())/1e6),
			Fifty:       fmt.Sprintf("%.2f", rec.RTTRecorder.Percentile(0.5)/1e6),
			SeventyFive: fmt.Sprintf("%.2f", rec.RTTRecorder.Percentile(0.75)/1e6),
			Ninety:      fmt.Sprintf("%.2f", rec.RTTRecorder.Percentile(0.90)/1e6),
		}
		if rec.UpCounter.Count() == 0 {
			item.RspRate = "0"
		} else {
			dt := float64(rec.DownCounter.Count()) / float64(rec.UpCounter.Count())
			item.RspRate = fmt.Sprintf("%.2f%%", dt*100)
		}
		mp[Mgr.handler.TransMsgId(key)] = item
		return true
	})
	return mp
}

func GetRecorder(k interface{}) *Recorder {
	val, ok := RecorderMap.Load(k)
	if ok {
		return val.(*Recorder)
	}

	rec := NewRecorder(k)
	val, _ = RecorderMap.LoadOrStore(k, rec)
	return val.(*Recorder)
}

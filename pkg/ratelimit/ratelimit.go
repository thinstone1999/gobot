package ratelimit

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type RateLimiter struct {
	tokens sync.Map
	conf   map[string]int32 // 协议名:每秒发送限制数量
}

// 设置上限
func SetLimit(conf map[string]int32) {
	limiter.conf = conf
	limiter.tokens = sync.Map{}
	limiter.Fill()
}

// 消费token
func Consume(name string) bool {
	return limiter.Consume(name)
}

// 消费token
func (rl *RateLimiter) Consume(name string) bool {
	v, ok := rl.tokens.Load(name)
	if ok {
		pv := v.(*int32)
		return atomic.AddInt32(pv, -1) >= 0
	}
	return true
}

// 填充token
func (rl *RateLimiter) Fill() {
	for name, limit := range rl.conf {
		count := limit
		if count < 0 {
			count = math.MaxInt32
		}
		rl.tokens.Store(name, &count)
	}
}

var limiter = &RateLimiter{}

func (rl *RateLimiter) loopFill() {
	tick := time.NewTicker(time.Second)
	for {
		rl.Fill()
		<-tick.C
	}
}

func init() {
	go limiter.loopFill()
}

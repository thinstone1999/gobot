package gobot

type Option func(o *Options)

// 配置项
type Options struct {
	StaticDir   string // 静态文件目录
	TreeFile    string // 行为树配置名
	TickMs      int64  // app tick
	StopWaitSec int64  // 退出前等待秒数
}

// 默认配置
const (
	AppName     = "nba"
	Height      = 720
	Width       = 1280
	StaticDir   = "static"
	TreeFile    = "robot.b3"
	TickMs      = 5000
	StopWaitSec = 0
)

func (op *Options) Apply(opts ...Option) {
	for i := range opts {
		opts[i](op)
	}
}

func WithStaticDir(dir string) Option {
	return func(o *Options) {
		o.StaticDir = dir
	}
}

func WithStopWaitSec(sec int64) Option {
	return func(o *Options) {
		o.StopWaitSec = sec
	}
}

func WithTreeFile(fname string) Option {
	return func(o *Options) {
		o.TreeFile = fname
	}
}

func WithTickMs(ms int64) Option {
	return func(o *Options) {
		o.TickMs = ms
	}
}

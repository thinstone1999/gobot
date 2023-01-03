package zap

import (
	"path/filepath"

	"go.uber.org/zap/zapcore"
)

type Options struct {
	Level      zapcore.Level
	Path       string
	Name       string
	Size       int32
	Stdout     bool
	ErrorAlone bool
}

type Option func(*Options)

func (opt *Options) Apply(ops ...Option) {
	for _, o := range ops {
		o(opt)
	}
}

func (opt *Options) getFname() string {
	if opt.Path == "" {
		return opt.Name + ".log"
	}

	return filepath.Join(opt.Path, opt.Name+".log")
}

func (opt *Options) getErrorname() string {
	if opt.Path == "" {
		return opt.Name + "_error.log"
	}

	return filepath.Join(opt.Path, opt.Name+"_error.log")
}

func EnableStdout() Option {
	return func(o *Options) {
		o.Stdout = true
	}
}

func WithErrorAlone() Option {
	return func(o *Options) {
		o.ErrorAlone = true
	}
}

func WithPath(path string) Option {
	return func(o *Options) {
		o.Path = path
	}
}

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func WithLevel(lv zapcore.Level) Option {
	return func(o *Options) {
		o.Level = lv
	}
}

func WithSize(path string) Option {
	return func(o *Options) {
		o.Path = path
	}
}

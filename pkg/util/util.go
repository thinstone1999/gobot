package util

import (
	"errors"
	"flag"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func Listen() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	return c
}

// 解析命令行参数到结构体
func ParseCmdArgs(v interface{}) error {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	if rt.Kind() != reflect.Ptr {
		return errors.New("mustBeStructPointer")
	}
	rt = rt.Elem()
	rv = rv.Elem()
	if rt.Kind() != reflect.Struct {
		return errors.New("mustBeStructPointer")
	}

	parseTag := func(tag string) (name, defaultVal, usage string) {
		list := strings.Split(tag, ",")
		n := len(list)
		if n > 0 {
			name = list[0]
		}
		if n > 1 {
			defaultVal = list[1]
		}
		if n > 2 {
			usage = list[2]
		}
		return
	}

	type intDesc struct {
		val int64
		fd  reflect.Value
	}
	dealList := []*intDesc{}
	defer func() {
		for _, item := range dealList {
			item.fd.SetInt(item.val)
		}
	}()

	for i := 0; i < rt.NumField(); i++ {
		fdt := rt.Field(i)
		fd := rv.Field(i)

		name, defaultVal, usage := parseTag(fdt.Tag.Get("arg"))
		if name == "" {
			name = strings.ToLower(fdt.Name)
		}

		switch fd.Kind() {
		case reflect.Int8, reflect.Int16,
			reflect.Int32, reflect.Int64, reflect.Int:
			item := &intDesc{
				fd: fd,
			}
			dealList = append(dealList, item)
			dflt, _ := strconv.ParseInt(defaultVal, 10, 64)
			flag.Int64Var(&item.val, name, dflt, usage)

		case reflect.Bool:
			pfdv := fd.Addr() // &fd
			pfdvnum := pfdv.Pointer()
			pfd := (*bool)(unsafe.Pointer(pfdvnum))
			dflt := false
			if defaultVal == "true" {
				dflt = true
			}
			flag.BoolVar(pfd, name, dflt, usage)

		case reflect.String:
			pfdv := fd.Addr() // &fd
			pfdvnum := pfdv.Pointer()
			pfd := (*string)(unsafe.Pointer(pfdvnum))
			flag.StringVar(pfd, name, defaultVal, usage)
		}

	}
	flag.Parse()
	return nil
}

func NowMs() int64 {
	return time.Now().UnixNano() / 1e6
}

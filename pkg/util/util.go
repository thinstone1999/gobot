package util

import (
	"encoding/base64"
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

// 需特殊处理的消息
type SpecialInfo struct {
	SpecialData map[string]interface{}
	DefaultData map[string]interface{}
	HasSpecial  bool
}

var excludeFieldMap = map[string]struct{}{
	"XXX_NoUnkeyedLiteral": {},
	"XXX_unrecognized":     {},
	"XXX_sizecache":        {},
}

// 获取结构体json默认值
func JsonDefault(tp reflect.Type) *SpecialInfo {
	// tp := reflect.TypeOf(v)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}

	if tp.Kind() != reflect.Struct {
		panic("supportStructOnly")
	}

	info := &SpecialInfo{}
	sp := doSpecial(tp, info)
	dft := doJsonDft(tp)
	info.SpecialData = sp.(map[string]interface{})
	info.DefaultData = dft.(map[string]interface{})
	return info
}

func doSpecial(tp reflect.Type, info *SpecialInfo) interface{} {
	for tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}

	switch tp.Kind() {
	case reflect.Struct:
		data := make(map[string]interface{})
		for i := 0; i < tp.NumField(); i++ {
			fd := tp.Field(i)
			if _, ok := excludeFieldMap[fd.Name]; !ok {
				tagName := strings.Split(fd.Tag.Get("json"), ",")[0]
				data[tagName] = doSpecial(fd.Type, info)
			}
		}
		return data

	case reflect.Slice:
		if tp.Elem().Kind() == reflect.Uint8 {
			info.HasSpecial = true
			return "bytes"
		}
		data := make([]interface{}, 1)
		data[0] = doSpecial(tp.Elem(), info)
		return data

	case reflect.Int64, reflect.Uint64:
		info.HasSpecial = true
		return "int64"
	}

	return reflect.New(tp).Interface()
}

func doJsonDft(tp reflect.Type) interface{} {
	for tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}

	switch tp.Kind() {
	case reflect.Struct:
		data := make(map[string]interface{})
		for i := 0; i < tp.NumField(); i++ {
			fd := tp.Field(i)
			if _, ok := excludeFieldMap[fd.Name]; !ok {
				tagName := strings.Split(fd.Tag.Get("json"), ",")[0]
				data[tagName] = doJsonDft(fd.Type)
			}
		}
		return data

	case reflect.Slice:
		if tp.Elem().Kind() == reflect.Uint8 {
			return ""
		}
		data := make([]interface{}, 1)
		data[0] = doJsonDft(tp.Elem())
		return data

	case reflect.Int64, reflect.Uint64:
		return ""
	}

	return reflect.New(tp).Interface()
}

// 转为可读的json
func EncodeJSONmap(v interface{}, dft interface{}) (bool, interface{}) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	rd := reflect.ValueOf(dft)
	if rd.Kind() == reflect.Ptr {
		rd = rd.Elem()
	}
	switch rv.Kind() {
	case reflect.Map:
		iter := rv.MapRange()

		for iter.Next() {
			// fmt.Printf("key: %v ", iter.Key().Interface())
			k := iter.Key().Interface()
			v := iter.Value().Interface()
			_, _ = k, v
			dv := rd.MapIndex(iter.Key()).Interface()
			if ok, val := EncodeJSONmap(iter.Value().Interface(), dv); ok {
				rv.SetMapIndex(iter.Key(), reflect.ValueOf(val))
				_ = val
			}
		}
		// fmt.Printf("}\n")
	case reflect.Slice:
		// fmt.Printf("\n[ ")
		dv := rd.Index(0).Interface()
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			if ok, val := EncodeJSONmap(elem.Interface(), dv); ok {
				elem.Set(reflect.ValueOf(val))
			}
		}
		// fmt.Printf("]\n")
	default:
		if rd.Kind() == reflect.String {
			switch rd.String() {
			case "int64":
				return true, strconv.Itoa(int(rv.Float()))

			case "bytes":
				data := rv.String()
				buff, _ := base64.StdEncoding.DecodeString(data)
				return true, string(buff)
			}
		}
		// fmt.Printf(" val: %v ", rv.Interface())
	}
	return false, nil
}

// 从可读json还原
func DecodeJSONmap(v interface{}, dft interface{}) (bool, interface{}) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	rd := reflect.ValueOf(dft)
	if rd.Kind() == reflect.Ptr {
		rd = rd.Elem()
	}
	switch rv.Kind() {
	case reflect.Map:
		iter := rv.MapRange()
		for iter.Next() {
			// fmt.Printf("key: %v ", iter.Key().Interface())
			k := iter.Key().Interface()
			v := iter.Value().Interface()
			_, _ = k, v
			dv := rd.MapIndex(iter.Key()).Interface()
			if ok, val := DecodeJSONmap(iter.Value().Interface(), dv); ok {
				rv.SetMapIndex(iter.Key(), reflect.ValueOf(val))
				_ = val
			}
		}
		// fmt.Printf("}\n")
	case reflect.Slice:
		// fmt.Printf("\n[ ")
		dv := rd.Index(0).Interface()
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			if ok, val := DecodeJSONmap(elem.Interface(), dv); ok {
				_ = val
			}
		}
		// fmt.Printf("]\n")
	default:
		if rd.Kind() == reflect.String {
			switch rd.String() {
			case "int64":
				data := rv.String()
				num, _ := strconv.ParseInt(data, 10, 64)
				return true, num

			case "bytes":
				data := rv.String()
				buff := base64.StdEncoding.EncodeToString([]byte(data))
				return true, buff
			}
		}
		// fmt.Printf(" val: %v ", rv.Interface())
	}
	return false, nil
}

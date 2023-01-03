package util

import "strconv"

type Map map[string]interface{}

func NewMap() Map {
	return Map{}
}

func (m Map) Set(k string, v interface{}) {
	m[k] = v
}

func (m Map) Get(k string) interface{} {
	return m[k]
}

func (m Map) GetString(k string) string {
	v, _ := m.Get(k).(string)
	return v
}

func (m Map) GetBool(k string) bool {
	v, _ := m.Get(k).(bool)
	return v
}

func (m Map) GetInt32(k string) int32 {
	return int32(m.GetInt64(k))
}

func (m Map) GetInt(k string) int {
	return int(m.GetInt64(k))
}

func (m Map) GetInt64(k string) int64 {
	v, ok := m[k]
	if !ok {
		return 0
	}

	switch typ := v.(type) {
	case int64:
		return typ
	case float64:
		return int64(typ)
	case float32:
		return int64(typ)
	case int:
		return int64(typ)
	case int32:
		return int64(typ)
	case string:
		num, _ := strconv.Atoi(typ)
		return int64(num)
	}
	return 0
}

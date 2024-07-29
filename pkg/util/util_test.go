package util

import (
	"testing"

	"github.com/Gonewithmyself/gobot/example/network/pb"
	myjson "github.com/Gonewithmyself/gobot/pkg/json"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestTypeDefault(t *testing.T) {
	// TypeDefault(&pb.)

	var m = map[string]int64{
		"a": myjson.MaxSafeInt + 1,
		"b": 1,
	}
	d, _ := myjson.Marshal(m)
	t.Log(string(d))
	var mm = make(map[string]int)
	myjson.Unmarshal(d, &mm)
	t.Log(mm)

	mt, err := protoregistry.GlobalTypes.FindMessageByName(proto.MessageName(&pb.GamerLocation{}))
	if err != nil {
		panic(err)
	}
	myjson.DisableOmitEmpty()
	d, _ = myjson.Marshal(mt.New().Interface())
	t.Log(string(d))
}

package util

import (
	"testing"

	"github.com/Gonewithmyself/gobot/example/network/pb"
	json "github.com/Gonewithmyself/gobot/pkg/myjson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestTypeDefault(t *testing.T) {
	// TypeDefault(&pb.)

	var m = map[string]int64{
		"a": json.MaxSafeInt + 1,
		"b": 1,
	}
	d, _ := json.Marshal(m)
	t.Log(string(d))
	var mm = make(map[string]int)
	json.Unmarshal(d, &mm)
	t.Log(mm)

	mt, err := protoregistry.GlobalTypes.FindMessageByName(proto.MessageName(&pb.GamerLocation{}))
	if err != nil {
		panic(err)
	}
	json.DisableOmitEmpty()
	d, _ = json.Marshal(mt.New().Interface())
	t.Log(string(d))
}

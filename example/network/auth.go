package network

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Gonewithmyself/gobot/example/types"

	"github.com/Gonewithmyself/gobot/back"
)

type BaseAuther struct {
	*types.AuthData
}

type Authreq struct {
	SdkName       string `json:"sdkName,omitempty"`
	SdkUid        string `json:"sdkUid,omitempty"`
	ClientVersion string `json:"clientVersion,omitempty"`
	ValidateInfo  string `json:"validateInfo,omitempty"`
	Channel       string `json:"channel,omitempty"`
	Device        string `json:"device,omitempty"`
}

type Authrsp struct {
	Code    int32  `json:"error"`
	Session string `json:"session"`
	UID     int64  `json:"uid"`
	Addr    string `json:"server"`
	// Zones         []*pb.LogicServerInfo `json:"zones"`
	LastLoginZone int32  `json:"lastzoneid"`
	Notice        string `json:"notice"`
	SdkUID        string
}

func (req Authreq) Data() string {
	d, _ := json.Marshal(req)
	return base64.StdEncoding.EncodeToString(d)
}

func NewAuther(auth *types.AuthData) back.IAuth {
	base := &BaseAuther{
		AuthData: auth,
	}

	return base
}

// sdk 认证
func (auth *BaseAuther) SdkAuth(cfg interface{}) (rsp interface{}, err error) {
	rsp = &Authreq{}
	return
}

func httpGet(url string) (rsp *http.Response, err error) {
	retry := 10
	for i := 0; i < retry; i++ {
		rsp, err = http.Get(url)
		if err == nil {
			return
		}
		time.Sleep(time.Millisecond * 50 * (time.Duration(i + 1)))
	}
	return
}

// 游戏服认证
func (auth *BaseAuther) GameAuth(reqinfo interface{}) (data interface{}, err error) {
	reqdata := reqinfo.(*Authreq)
	url := fmt.Sprintf("%v?data=%v", auth.Conf.URL, reqdata.Data())
	resp, err := httpGet(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var info Authrsp
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return
	}
	info.SdkUID = reqdata.SdkUid
	data = &info
	return
}

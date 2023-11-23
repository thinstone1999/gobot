package back

// 认证接口
type IAuth interface {
	SdkAuth(cfg interface{}) (interface{}, error)     // sdk认证
	GameAuth(sdkrsp interface{}) (interface{}, error) // 游戏服认证
}

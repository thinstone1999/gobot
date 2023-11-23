package types

type (
	// sdk 配置信息
	SdkConfig struct {
		Name      string `json:"name,omitempty"`
		AppID     string `json:"app_id,omitempty"`
		AppKey    string `json:"app_key,omitempty"`
		Secret    byte   `json:"secret"`
		ChannelID string `json:"channel_id,omitempty"`
		URL       string `json:"url,omitempty"`            // SDK地址
		Prefix    string `json:"account_prefix,omitempty"` // 前缀 + 序号 作为账号
	}

	// 游戏服 登录配置
	LoginConfig struct {
		URL           string      `json:"url,omitempty"` // 认证服地址
		Zone          int32       `json:"zone,omitempty"`
		Device        interface{} `json:"device_id,omitempty"`
		Account       interface{} `json:"account,omitempty"` // 压测时忽略
		ClientVersion string      `json:"client_version,omitempty"`
	}

	AuthData struct {
		Conf       *LoginConfig
		SdkAccount string
	}

	// 命令行参数
	CmdArgs struct {
		Tree     string `arg:",01,压测用例"`
		Start    int32  `arg:"start,1,起始序号"`
		Count    int32  `arg:"count,1,压测数量"`
		Timeout  int32  `arg:"timeout,0,压测时间(秒)"`
		Zone     int32  `arg:"zone,1,区服"`
		Auth     string `arg:"auth,http://localhost:5000/authtoken,认证服地址"`
		StopWait int32  `arg:"wait,0,压测停止后等待时间(秒)"`
	}

	AppConfig struct {
		LogLevel  string       `json:"log_level,omitempty"`
		LogPath   string       `json:"log_path,omitempty"`
		TickMs    int64        `json:"tick_ms,omitempty"` // 行为树每x毫秒跑一次
		SdkConf   *SdkConfig   `json:"sdk_conf,omitempty"`
		LoginConf *LoginConfig `json:"login_conf,omitempty"`
	}
)

var (
	SdkConf *SdkConfig
	Args    CmdArgs
	AppConf *AppConfig
)

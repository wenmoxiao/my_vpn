package mode
import(
	_"encoding/json"
	_"io/ioutil"
	"io/ioutil"
	"encoding/json"
)


const(
	OK_CONECT_SERVER  	 = 100		//socket连接到服务器成功
	ERROR_CONECT_SERVER  = 101		//socket连接到服务器失败
)


type SocketConfig struct{
	m_version  byte
	m_connect  byte
}

type Upstream struct {
	Type     string `json:"type"`
	Crypto   string `json:"crypto"`
	Password string `json:"password"`
	Address  string `json:"address"`
}

type PAC struct {
	Address     string   `json:"address"`
	Proxy       string   `json:"proxy"`
	SOCKS5      string   `json:"socks5"`
	LocalRules  string   `json:"local_rule_file"`
	RemoteRules string   `json:"remote_rule_file"`
	Upstream    Upstream `json:"upstream"`
}

type Proxy struct {
	HTTP            string     `json:"http"`
	SOCKS4          string     `json:"socks4"`
	SOCKS5          string     `json:"socks5"`
	Crypto          string     `json:"crypto"`
	Password        string     `json:"password"`
	DNSCacheTimeout int        `json:"dnsCacheTimeout"`
	Upstreams       []Upstream `json:"upstreams"`
}



type Config struct {
	PAC     PAC     `json:"pac"`
	Proxies []Proxy `json:"proxies"`
}

type ServerConfig struct{
	ConnMax				int		`json:"connMax"`				// 连接最大数
	NetType		  		string  `json:"tcp"`					// 连接类型
	ListenAddress 		string  `json:"address"`				// 监听地址
	LogFile       		string  `json:"logfile"`				// 日志文件名
	DnsCacheTime 		int     `json:"dnsCacheTime"`			// dns缓存时间
	FwdAddress			string  `json:"fwdAddress"`				// 回源地址
	EType				string  `json:"eType"`					// 加密方式
	EPassword			string  `json:"ePassword"`				// 加密密钥
	SocketType			string  `json:"socketType"`				// socket类型 socket4 socket5
}

func LoadConfig(s string) (*Config, error) {
	data, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	cfgGroup := &Config{}
	if err = json.Unmarshal(data, cfgGroup); err != nil {
		return nil, err
	}
	return cfgGroup, nil
}

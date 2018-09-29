package mode

import (
	"net"
	"io"
	"strings"
	"00_my_vpn/src/common/Encry"
	"00_my_vpn/src/common/CLog"
)
/*=====================================================================================================*/
// 连接对象
/*=====================================================================================================*/
type SConn struct{
	net.Conn							// 继承conn接口
	rwc 		io.ReadWriteCloser		// 继承流传递接口
}

func (c *SConn) Read(data []byte) (int, error) {
	return c.rwc.Read(data)
}
func (c *SConn) Write(data []byte) (int, error) {
	return c.rwc.Write(data)
}
func (c *SConn) Close() error {
	err := c.Conn.Close()
	c.rwc.Close()
	return err
}

// 创建连接
func NewConnent(conn net.Conn, cryptMethod string, password []byte) (*SConn, error){
	var rwc io.ReadWriteCloser
	var err error
	switch strings.ToLower(cryptMethod) {
	default:
		rwc = conn
	case "rc4":
		rwc, err = Encry.NewRC4Cipher(conn, password)
	case "des":
		rwc, err = Encry.NewDESCFBCipher(conn, password)
	case "aes-128-cfb":
		rwc, err = Encry.NewAESCFGCipher(conn, password, 16)
	case "aes-192-cfb":
		rwc, err = Encry.NewAESCFGCipher(conn, password, 24)
	case "aes-256-cfb":
		rwc, err = Encry.NewAESCFGCipher(conn, password, 32)
	}
	if err != nil {
		return nil, err
	}

	return &SConn{
		Conn: conn,
		rwc:  rwc,
	}, nil
}


/*=====================================================================================================*/
// 连接对象管理器
/*=====================================================================================================*/
type SConnManager struct {
	conn 		map[string]SConn		// 连接对象列表
	DnsM			*DNSManager			// Dns管理器
	log 			*Clog.C_LOG			// log
}

// 创建管理器
func NewSConnManager(conn_max int, dnsCacheTime int, out_log *Clog.C_LOG) *SConnManager {
	return  &SConnManager{
        conn : make(map[string]SConn,conn_max),
		DnsM : NewDNSManager(dnsCacheTime),
		log : out_log,
	}
}

// 销毁管理器
func (th *SConnManager)Distory() int{
    th.log = nil
    th.DnsM = nil
	return 0
}

// 解析地址
func (th *SConnManager)parseAddress(address string) (interface{}, string, error) {
	host, port, err := net.SplitHostPort(address) // 解析 ip port
	if err != nil {
		return nil, "", err
	}
	ip := net.ParseIP(address)
	if ip != nil {
		return ip, port, nil
	} else {
		return host, port, nil
	}
}

// 获取连接
func (th *SConnManager)GetConn(network, address string, cryptMethod string, password string) (*SConn, error){
	// 解析ip
	host, port, err := th.parseAddress(address)
	if err != nil {
		th.log.Print(Clog.L_ERROR,"GetConn address is error ", address, " ", err)
		return nil, err
	}
	var dest string
	var ipCached bool
	switch h := host.(type) {  // 获取host实际的类型
	case net.IP:              // ip类型
		{
			dest = h.String()
			ipCached = true
		}
	case string:             // 域名类型
		{
			dest = h
			if th.DnsM != nil {          // 在dns缓存中查找
				if p, ok := th.DnsM.Get(h); ok {
					dest = p.String()
					ipCached = true
				}
			}
		}
	}
	th.log.Print(Clog.L_INFO,"GetConn ip_info  ", address, " ", dest," ", ipCached)
	// 生成连接
	address = net.JoinHostPort(dest, port)  // 生成新的ip地址
	destConn, err := net.Dial(network, address) // 连接
	if err != nil {
		th.log.Print(Clog.L_ERROR,"GetConn conn error ", address, " ", dest," ", ipCached, " ", err)
		return nil, err
	}
	if th.DnsM != nil && !ipCached { // 更新dns信息
		th.DnsM.Set(host.(string), destConn.RemoteAddr().(*net.TCPAddr).IP)
	}

	// 建立流处理连接器
	sConn, err := NewConnent(destConn, cryptMethod, []byte(password))
	if err != nil {
		th.log.Print(Clog.L_ERROR,"GetConn get SConn error ", address, " ", dest," ", ipCached, " ", err)
		destConn.Close()
		return nil, err
	}

	th.log.Print(Clog.L_INFO,"GetConn conn success ", address, " ", dest," ", ipCached, " ", &destConn)
	return sConn, nil
}
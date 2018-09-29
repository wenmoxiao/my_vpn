package mode


import (
	_ "fmt"
	_ "io"
	_ "errors"
	_"net"
	_"net/url"
	"net"
	"00_my_vpn/src/common/CLog"
)


// 定义接口对象
type SocketServer struct {
	// 运行前生效
	conf 			*ServerConfig					// 配置
	log 			*Clog.C_LOG						// log
	SConnM			*SConnManager					// 连接句柄管理器
	// 运行期生效
	Listener 		net.Listener					// 监听句柄
}

/*=====================================================================================================*/
//  实现接口函数
/*=====================================================================================================*/
// 继承自SServer
func (th *SocketServer)Init(obj []interface{}) int{
	if len(obj) <= 0{
		return -1
	}
	var inconfig *ServerConfig
	switch obj[0].(type){
	case *ServerConfig:
		inconfig = obj[0].(*ServerConfig)
		break
	default:
		return -2
	}
	th.conf = inconfig
	th.Listener = nil
	th.log = Clog.Create(inconfig.LogFile, 0)
	th.SConnM = NewSConnManager(inconfig.ConnMax, inconfig.DnsCacheTime, th.log)
	th.log.Print(Clog.L_INFO,"http server init success! ")
	return 0
}

func (th *SocketServer) Begin() int {
	// 监听
	if th.conf.ListenAddress != ""{
		var err error
		th.Listener, err =net.Listen(th.conf.NetType, th.conf.ListenAddress)
		if err != nil {
			th.log.Print(Clog.L_ERROR, "listen address error , net_type:", th.conf.NetType, " address:", th.conf.ListenAddress, " error:",err)
			return -1
		}
	}

	th.log.Print(Clog.L_INFO,"http server Begin success! ", &th.Listener)
	return 0
}
func (th *SocketServer) Run() int {
	// 这个要自己实现
	for {
		conn, err := th.Listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else {
				return -1
			}
		}
		// 这里需要建立client自定义连接
		sconn, err := NewConnent(conn,th.conf.EType, []byte(th.conf.EPassword))
		if err != nil {
			return -2
		}

		if  th.conf.SocketType == "socket4"{
			go serveSOCKS4Client(sconn, th) // 独立出一个线程处理
		}
		if th.conf.SocketType == "socket5" {
			go serveSocks5Client(sconn, th) // 独立出一个线程处理
		}

	}
	return 0
}

func (th *SocketServer) End() int {
	if nil != th.Listener{
		th.Listener.Close()
	}
	return 0
}

func (th *SocketServer) Close() int {
	th.conf=nil
	th.log=nil
	th.SConnM.Distory()
	th.SConnM = nil
	return 0
}

/*=====================================================================================================*/
// 实现回源处理部分
/*=====================================================================================================*/
func (th *SocketServer)Dial(network, address string)(net.Conn, error){
	// 这里需要通过配置指定的回源ip建立连接
	conn, err := th.SConnM.GetConn(network, address,th.conf.EType, th.conf.EPassword)
	if err != nil {
		return nil, err
	}
	return conn,nil
}
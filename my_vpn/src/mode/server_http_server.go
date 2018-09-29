package mode

import (
	_ "fmt"
	_ "io"
	_ "errors"
	_"net"
	_"net/http"
	_"net/http/httputil"
	_"net/url"
	"net"
	"fmt"
	"io"
	"net/http/httputil"
	"net/http"
	"net/url"
	"00_my_vpn/src/common/CLog"
	"strconv"
	"errors"
)


// 定义接口对象
type HTTPServer struct {
	// 运行前生效
	conf 			*ServerConfig					// 配置
	log 			*Clog.C_LOG						// log
	ReverseProxy 	*httputil.ReverseProxy 			// http 反向代理句柄
	SConnM			*SConnManager					// 连接句柄管理器
	// 运行期生效
	Listener 		net.Listener					// 监听句柄
}

/*=====================================================================================================*/
//  实现接口函数
/*=====================================================================================================*/
// 继承自SServer
func (th *HTTPServer)Init(obj []interface{}) int{
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
	// 构造反向代理对象
	th.ReverseProxy = &httputil.ReverseProxy{
		Director: th.HTTPDirector,
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return th.Dial(network, addr)
			},
		},
	}
	th.log.Print(Clog.L_INFO,"http server init success! ", &th.ReverseProxy)
	return 0
}

func (th *HTTPServer) Begin() int {
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
func (th *HTTPServer) Run() int {
	http.Serve(th.Listener, th)
	return 0
}

func (th *HTTPServer) End() int {
	if nil != th.Listener{
		th.Listener.Close()
	}
	return 0
}

func (th *HTTPServer) Close() int {
	th.conf=nil
	th.log=nil
	th.SConnM.Distory()
	th.SConnM = nil
	th.ReverseProxy=nil
	return 0
}


/*=====================================================================================================*/
// 补充功能函数
/*=====================================================================================================*/
// 反向代理挂载点,用于修改request属性; httprequest处理器
func (th *HTTPServer)HTTPDirector(request *http.Request) {
	th.log.Print(Clog.L_INFO,"ServeHTTP_HTTPDirector ", request.RequestURI)
	u, err := url.Parse(request.RequestURI)
	if err != nil {
		th.log.Print(Clog.L_ERROR,"ServeHTTP_HTTPDirector url error ", request.RequestURI," ", err)
		return
	}
	// 重新修改url
	request.RequestURI = "http:/" + u.RequestURI()
	// 这里修改连接行为
	v := request.Header.Get("Proxy-Connection")
	if v != "" {
		request.Header.Del("Proxy-Connection")
		request.Header.Del("Connection")
		request.Header.Add("Connection", v)
	}

}

// 已连接下的隧道处理
func (th *HTTPServer)HTTPTunnel(response http.ResponseWriter, request *http.Request) int{
	th.log.Print(Clog.L_INFO,"HTTPTunnel start ", request.Method, " ", request.RequestURI)
	// 客户端连接
	var conn net.Conn
	if hj, ok := response.(http.Hijacker); ok {
		var err error
		if conn, _, err = hj.Hijack(); err != nil {
			th.log.Print(Clog.L_INFO,"HTTPTunnel response ", request.RequestURI, " 500 ", err.Error())
			http.Error(response, err.Error(), http.StatusInternalServerError) // 返回给客户端500
			return -1
		}
	} else {
		th.log.Print(Clog.L_INFO,"HTTPTunnel response ", request.RequestURI, "  500 Hijacker failed")
		http.Error(response, "Hijacker failed", http.StatusInternalServerError)
		return -2
	}
	defer conn.Close()

	// server 连接
	dest, err := th.Dial("tcp", request.Host)
	if err != nil {
		th.log.Print(Clog.L_INFO,"HTTPTunnel fwd conn error ", request.RequestURI, "  500 Hijacker failed")
		fmt.Fprintf(conn, "HTTP/1.0 500 NewRemoteSocks failed, err:%s\r\n\r\n", err)
		return -3
	}
	defer dest.Close()

	// 将客户端请求的信息发送到目标server
	if request.Body != nil {
		if _, err = io.Copy(dest, request.Body); err != nil {
			fmt.Fprintf(conn, "%d %s", http.StatusBadGateway, err.Error())
			return -4
		}
	}
	// 构建返回头部
	fmt.Fprintf(conn, "HTTP/1.0 200 Connection established\r\n\r\n")

	go func() {
		defer conn.Close()
		defer dest.Close()
		th.log.Print(Clog.L_INFO,"HTTPTunnel send info:")
		io.Copy(dest, conn)  // 将客户端的数据拷贝到源站
	}()

	data  := make([]byte, 1024*1000)
	dest.Read(data)
	th.log.Print(Clog.L_INFO,"HTTPTunnel send info1:",data )
	io.Copy(conn, dest) // 拷贝源站的数据到客户端
	return 0
}


/*=====================================================================================================*/
// 实现回源处理部分(net.http 关键部分)
/*=====================================================================================================*/
// http 用于外部自定义主入口的函数; 继承自HTTP Handler
func (th *HTTPServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	th.log.Print(Clog.L_INFO,"ServeHTTP new start ", request.Method, " ", request.RequestURI)
	th.HTTPTunnel(response, request)
	/*
	if request.Method == "CONNECT" {
		th.HTTPTunnel(response, request)
	} else {
	//	th.ReverseProxy.ServeHTTP(response, request) // 丢给原始的入口操作
	}
	*/
}


// 拨号连接接口(可以通过实现这个提前准备数据) 回源连接建立
func (th *HTTPServer)Dial(network, address string)(net.Conn, error){
	th.log.Print(Clog.L_INFO,"ServeHTTP_Dial new conn start ", address)
	// 检查连接类型
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		th.log.Print(Clog.L_ERROR,"ServeHTTP_Dial network error type:" + network)
		return nil, errors.New("ServeHTTP_Dial network error type:" + network)
	}

	// 分解host 和端口
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		th.log.Print(Clog.L_ERROR,"ServeHTTP_Dial host or port error address:" + address)
		return nil, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		th.log.Print(Clog.L_ERROR,"ServeHTTP_Dial port error portStr:" + portStr)
		return nil, errors.New("ServeHTTP_Dial port error portStr" + portStr)
	}
	if port < 1 || port > 0xffff {
		th.log.Print(Clog.L_ERROR,"ServeHTTP_Dial port error range:" + portStr)
		return nil, errors.New("ServeHTTP_Dial port error range:" + portStr)
	}

	// 这里需要通过配置指定的回源ip建立连接
	conn, err := th.SConnM.GetConn(th.conf.NetType, th.conf.FwdAddress,th.conf.EType, th.conf.EPassword)
	if err != nil {
		return nil, err
	}

	closeConn := &conn
	defer func() {
		if closeConn != nil {
			(*closeConn).Close()
		}
	}()

	// 截取host
	buff := make([]byte, 0, 266)
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			buff = append(buff, 1)
			ip = ip4
		} else {
			buff = append(buff, 4)
		}
		buff = append(buff, ip...)
	} else {
		if len(host) > 255 {
			return nil, errors.New("socks: destination hostname too long: " + host)
		}
		buff = append(buff, 3)
		buff = append(buff, uint8(len(host)))
		buff = append(buff, host...)
	}
	buff = append(buff, uint8(port>>8), uint8(port))

	th.log.Print(Clog.L_INFO,"ServeHTTP_Dial send info:" ,buff)
	// 发送给目标服务器
	_, err = conn.Write(buff)
	if err != nil {
		return nil, err
	}

	closeConn = nil
	return conn, nil
}










package mode

import(
	_"net"
	"net"
)
// 接口类
// 服务接口类
type SServer interface{
	// 服务框架
	Init(obj []interface{}) int      					// 服务初始化
	Begin() int											// 服务开启阶段
	Run() int											// 服务执行阶段
	End() int											// 服务结束阶段
	Close() int											// 服务清理
	// 其他
	Dial(network, address string)(net.Conn, error)		// server 连接函数(连接拨号用)
}


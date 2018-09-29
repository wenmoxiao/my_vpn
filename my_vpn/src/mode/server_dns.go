package mode

import (
	"net"
	"sync"
	"time"
)

type DNSElement struct {
	Ip        		net.IP								// ip信息
	StartTime 		time.Time							// 开始缓存时间
}

type DNSManager struct {
	Lock         	sync.RWMutex						// 互斥锁
	Timeout 		time.Duration						// 缓存时间
	Dns          	map[string]DNSElement				// dns元素
}


/*=====================================================================================================*/
// DNS 缓存部分
/*=====================================================================================================*/
var GSDNSManager *DNSManager = nil

// 创建dns 管理器
func NewDNSManager(timeout int) *DNSManager {
	if GSDNSManager != nil {
		return GSDNSManager
	}
	if timeout <= 0 {
		timeout = 30
	}
	GSDNSManager= &DNSManager{
		Timeout: time.Duration(timeout) * time.Minute,
		Dns:     make(map[string]DNSElement),
	}
	return GSDNSManager
}

// 根据url 获取会员ip
func (th *DNSManager) Get(key string) (net.IP, bool) {
	th.Lock.RLock()
	e, ok := th.Dns[key]
	th.Lock.RUnlock()
	// 超时
	if ok && time.Since(e.StartTime) > th.Timeout {
		th.Lock.Lock()
		delete(th.Dns, key)
		th.Lock.Unlock()
		return nil, false
	}
	return e.Ip, ok
}

// 本地缓存路由信息
func (th *DNSManager) Set(key string, ip net.IP) int{
	th.Lock.Lock()
	th.Dns[key] = DNSElement{Ip: ip, StartTime: time.Now()}
	th.Lock.Unlock()
	return 0
}

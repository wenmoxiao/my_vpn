package main

import (
	_ "fmt"
	_ "io/ioutil"
	_ "net/http"
	_ "net/url"
	_ "encoding/json"
	_ "00_my_vpn/src/common"
	_ "00_my_vpn/src/common/CLog"
	_ "00_my_vpn/src/mode"
	"00_my_vpn/src/common/CLog"
	"00_my_vpn/src/mode"
)


/*
// run porxy server
func runHTTPPoryServer(conf Proxy, router socks.Dialer) {

}
*/

var TManager *ThreadManager

// log 相关
var thislog *Clog.C_LOG
func InitLog(){
	Clog.Init(10)
	thislog = Clog.Create("txt.log", 0)
}
func CloseLog(){
	Clog.Clear()
}


// 线程调度相关
type ThreadInfo struct{
	Block 			chan int
	pid int
}
func (th *ThreadInfo)Init(MS mode.SServer,obj []interface{}) int{
	MS.Init(obj)
	return 0
}

func (th *ThreadInfo)Run(MS mode.SServer) int{
	MS.Begin()
  go func(){
    MS.Run()
    MS.End()
  	th.Block <- 0
  }()

  return 0
}

func (th *ThreadInfo)Close() int{
   <- th.Block
   return 0
}


type ThreadManager struct{
	MaxWorker 		int
	CanCreateMax  int
	TInfo 			[]*ThreadInfo
}

func CreateTheadManager(createMax int) *ThreadManager{
    TM := new(ThreadManager)
	TM.MaxWorker = 0
	TM.CanCreateMax = createMax
	return TM
}

func (TM *ThreadManager)CreateThread(MS mode.SServer, obj ...interface{}) *ThreadInfo{
  if TM.MaxWorker >= TM.CanCreateMax {
  	return nil
  }
  TI := new(ThreadInfo)
  TM.TInfo =append(TM.TInfo, TI)  // 追加到数组
  TM.MaxWorker++
  TI.Init(MS, obj)
  TI.Run(MS)
  return TI
}

func (TM *ThreadManager)CloseThreadManager() int{
  for i := 0; i < TM.MaxWorker; i++{
  	TM.TInfo[i].Close()
  }
  return 0
}





func main() {

	InitLog()
	TManager = CreateTheadManager(5)
	// 加载配置
	hpc := &mode.ServerConfig{
		ConnMax: 20000,
		NetType : "tcp",
		ListenAddress :":4000",
		LogFile :"http_log.txt",
		DnsCacheTime : 30,
		FwdAddress : "baidu.com:80",
		EType : "",
		EPassword : "",
		SocketType : "",
	}


	// http server
	httpS := new(mode.HTTPServer)
	TManager.CreateThread(httpS, hpc)


	TManager.CloseThreadManager()
	CloseLog()
}
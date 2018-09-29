package Clog

import (
	_ "log"
	_ "os"
	"log"
)

type S_LOG []*C_LOG

type LOG_MABAGER struct {
	log_list S_LOG
	max_num int
	cur_num int
}
var LOG_M = LOG_MABAGER{}

// 插入元素
func Insert(l *C_LOG) int{
	if LOG_M.cur_num + 1 >= LOG_M.max_num{
		log.Println("log manager error")
		return -1
	}
	LOG_M.log_list[LOG_M.cur_num] = l
	LOG_M.cur_num++
	return 0
}


// 初始化管理器
func Init(num int) {
	LOG_M.log_list = make(S_LOG, num)
	LOG_M.max_num = num
	LOG_M.cur_num = 0
	for i := 0; i < num; i++{
		LOG_M.log_list[i] = nil
	}
}

// 清理管理器
func Clear(){
   for i := 0; i < LOG_M.max_num; i++{
   	if LOG_M.log_list[i] != nil {
		LOG_M.log_list[i].Delete()
	}
   }
   LOG_M.log_list = nil
   LOG_M.max_num = 0
   LOG_M.cur_num = 0
}


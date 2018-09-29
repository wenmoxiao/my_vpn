package Clog

import (
	_ "log"
	_ "os"
	"log"
	"os"
)

const (
	L_DEBUG int = iota 	// value --> 0
	L_INFO              // value --> 1
	L_ERROR           	// value --> 2
	L_MAX           	// value --> 3
)


type C_LOG struct{
	my_log *log.Logger
	level int
	log_file *os.File
}

func GetLogLevel(level int) string{
	switch level{
	case L_DEBUG:
		return "[Debug]"
	case L_INFO:
		return "[Info]"
	case L_ERROR:
		return "[Error]"
	}
	return "[Debug]"
}

func Create(file_name string, level int) *C_LOG{
	this_file, this_err := os.Create(file_name)
	if this_err != nil{
		log.Println("open log file error ",file_name)
		return nil
	}
	var tt *C_LOG
	tt = &C_LOG{
     	my_log: log.New(this_file, "\r\n" + GetLogLevel(level),log.Ldate | log.Ltime),
     	level: level,
     	log_file: this_file,
	 }
	Insert(tt)
	return tt
}

func (h *C_LOG)Delete(){
     h.log_file.Close()
}

func (h *C_LOG)Print(level int,v ...interface{}){
	 if level != h.level{
		 h.my_log.SetPrefix(GetLogLevel(level))
	 }
	//_, file, line, ok = runtime.Caller(1)
     h.my_log.Println(v)
}


package main

import (
	"github.com/gysan/chatroom/config"
	"github.com/gysan/chatroom/dao"
	"github.com/gysan/chatroom/handlers"
	"github.com/gysan/chatroom/server"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"flag"
	"github.com/golang/glog"
	"github.com/gysan/chatroom/common"
)

var svr *server.Server

func init() {
	// 读取配置文件
	err := config.ReadIniFile("./config.ini")
	checkError(err)

	// 设置cpu数量和日志目录
	runtime.GOMAXPROCS(runtime.NumCPU())
	setLogOutput(config.LogFile)

	// 初始化dao
	dao.IdMsgInit(config.IdToMsgDB)
	dao.OfflineMsgInit(config.OfflineMsgidsDB)
	dao.UuidInit(config.UuidDB)

	// 服务器初始化
	svr = server.NewServer()
	svr.SetAcceptTimeout(time.Duration(config.AcceptTimeout) * time.Second)
	svr.SetReadTimeout(time.Duration(config.ReadTimeout) * time.Second)
	svr.SetWriteTimeout(time.Duration(config.WriteTimeout) * time.Second)

	// 消息处理函数绑定
	svr.BindMsgHandler(uint32(common.MessageCommand_HEARTBEAT_INIT), handlers.HandleHeartbeatInit)
	svr.BindMsgHandler(uint32(common.MessageCommand_HEARTBEAT), handlers.HandleHeartbeat)
	svr.BindMsgHandler(uint32(common.MessageCommand_USER_LOGIN), handlers.HandleUserLogin)
	svr.BindMsgHandler(uint32(common.MessageCommand_USER_LOGOUT), handlers.HandleUserLogout)
	svr.BindMsgHandler(uint32(common.MessageCommand_MESSAGE), handlers.HandleMessage)
	svr.BindMsgHandler(uint32(common.MessageCommand_RECEIVE_MESSAGE_ACK), handlers.HandleReceiveMessageAck)
	svr.BindMsgHandler(uint32(common.MessageCommand_NORMARL_MESSAGE_ACK), handlers.HandleNormalMessageAck)
}

func clean() {
	glog.Info("DB cleaning...")
	dao.IdMsgClean()
	dao.OfflineMsgClean()
	dao.UuidClean()
	glog.Info("DB cleaned.")
}

func main() {
	flag.Parse()
	glog.Info("IM start...")
	glog.Flush()

	tcpAddr, err := net.ResolveTCPAddr("tcp4", config.Addr)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	// 服务器开始监听
	go svr.Start(listener)
	go server.StartHttp()
	// 处理中断信号
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	handleSignal(ch)

	// 优雅的结束线程
	svr.Stop()
	clean()
	glog.Info("IM stop.")
}

func checkError(err error) {
	if err != nil {
		glog.Errorf("Error: %v", err)
		os.Exit(1)
	}
}

func setLogOutput(filepath string) {
	// 为log添加短文件名，方便查看行数
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	logfile, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	// 注意不要关闭logfile
	if err != nil {
		log.Printf("%v\r\n", err)
	}
	log.SetOutput(logfile)
}

func handleSignal(c chan os.Signal) {
	glog.Info("HandleSignal start.")
	// Block until a signal is received.
	for {
		s := <-c
		glog.Infof("comet get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			glog.Info("Signal quit, term, stop, int.")
			return
		case syscall.SIGHUP:
			glog.Info("Signal hup.")
			// TODO reload
			//return
		default:
			glog.Info("Signal default.")
			return
		}
	}
	glog.Info("HandleSignal end.")
}

package main

import (
	proto "code.google.com/p/goprotobuf/proto"
	"flag"
	"fmt"
	"github.com/gysan/chatroom/config"
	"github.com/gysan/chatroom/handlers"
	"github.com/gysan/chatroom/packet"
	"github.com/gysan/chatroom/utils/convert"
	"net"
	"os"
	"time"
	"github.com/gysan/chatroom/common"
	"github.com/golang/glog"
	"strconv"
)

var (
	sender string
	receiver string

	heartbeatDelay int32 = 20
	tag uint32           = 1
)

func getPacFromBuf(buf []byte, n int) *packet.Packet {
	pacLen := convert.BytesToUint32(buf[8:12])
	pac := &packet.Packet{
		Tag: convert.BytesToUint32(buf[0:4]),
		Type: convert.BytesToUint32(buf[4:8]),
		Length:  pacLen,
		Data: buf[12:int(pacLen) + 12],
	}

	if n != int(pacLen)+12 {
		glog.Infof("Tag: %v, Type: %v, Length: %v, Data: %v", pac.Tag, pac.Type, pac.Length, pac.Data)
		return nil
	}
	glog.Infof("Buffer %v to packet: %v", buf, pac)
	return pac
}

// 发送心跳包
func ping(conn *net.TCPConn) {
	glog.Info("Heartbeat beginning...")
	ticker := time.NewTicker(time.Duration(heartbeatDelay) * time.Second)
	for _ = range ticker.C {
		heartbeat := &common.Heartbeat{
			LastDelay: proto.Int32(heartbeatDelay),
		}
		err := handlers.SendPbData(conn, tag, uint32(common.MessageCommand_HEARTBEAT), heartbeat)
		if err != nil {
			glog.Errorf("Send heartbeat data error: %v", err)
			return
		}
		glog.Infof("Heartbeat: %v", conn.RemoteAddr().String())
	}
	glog.Info("Heartbeat end.")
}

func heartbeatInit(conn *net.TCPConn) {
	glog.Infof("HeartbeatInit...")
	heartbeatInit := &common.HeartbeatInit{
		LastTimeout: proto.Int32(heartbeatDelay),
	}
	glog.Infof("Heartbeat init: %v", heartbeatInit)
	err := handlers.SendPbData(conn, tag, uint32(common.MessageCommand_HEARTBEAT_INIT), heartbeatInit)
	if err != nil {
		glog.Errorf("Send data error: %v", err)
		return
	}

	pack, err := packet.GetPacketFromBuffer(conn)
	if err != nil {
		glog.Errorf("Packet buffer: %v", err)
		return
	}
	tag = pack.Tag+1
	heartbeatInitResponse := &common.HeartbeatInitResponse{}
	err = packet.Unpack(pack, heartbeatInitResponse)
	if err != nil {
		glog.Errorf("%v Unpack error: %v\r\n", sender, err)
		return
	}
	glog.Infof("HeartbeatInitResponse: %v", heartbeatInitResponse)
}

func userLogin(conn *net.TCPConn, sender string) {
	glog.Infof("UserLogin...")
	userLogin := &common.UserLogin{
		UserId: proto.String(sender),
	}
	glog.Infof("UserLogin: %v", userLogin)
	err := handlers.SendPbData(conn, tag, uint32(common.MessageCommand_USER_LOGIN), userLogin)
	if err != nil {
		glog.Errorf("Send data error: %v", err)
		return
	}

	pack, err := packet.GetPacketFromBuffer(conn)
	if err != nil {
		glog.Errorf("Packet buffer: %v", err)
		return
	}
	tag = pack.Tag+1
	userLoginResponse := &common.UserLoginResponse{}
	err = packet.Unpack(pack, userLoginResponse)
	if err != nil {
		glog.Errorf("%v Unpack error: %v\r\n", sender, err)
		return
	}
	glog.Infof("UserLoginResponse: %v", userLoginResponse)
}

func userLogout(conn *net.TCPConn, sender string) {
	glog.Infof("UserLogout...")
	userLogout := &common.UserLogout{
		UserId: proto.String(sender),
	}
	glog.Infof("UserLogout: %v", userLogout)
	err := handlers.SendPbData(conn, tag, uint32(common.MessageCommand_USER_LOGOUT), userLogout)
	if err != nil {
		glog.Errorf("Send data error: %v", err)
		return
	}
}

func sendMessage(conn *net.TCPConn, sender string, receiver string) {
	ticker := time.NewTicker(time.Duration(heartbeatDelay+10) * time.Second)
	for _ = range ticker.C {
		glog.Infof("%v send message to %v ...", sender, receiver)
		message := &common.Message{
			Sender: proto.String(sender),
			Receiver: proto.String(receiver),
			MessageId: proto.String(sender + "_" + receiver + "_" + strconv.Itoa(int(time.Now().Unix()))),
			MessageBody: proto.String("{\"type\":\"1\", \"text\":\"Hello world\"}"),
		}
		glog.Infof("Message: %v", message)
		err := handlers.SendPbData(conn, tag, uint32(common.MessageCommand_MESSAGE), message)
		if err != nil {
			glog.Errorf("Send data error: %v", err)
			return
		}
	}
	glog.Info("SendMessage end.")
}

func testBB(sender string) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", config.Addr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		glog.Infof("%v DialTCP error: %v", sender, err)
		return
	}
	defer conn.Close()

	heartbeatInit(conn)

	userLogin(conn, sender)

	// 定时发送心跳包
	go ping(conn)

	go sendMessage(conn, sender, receiver)

	// 死循环，接收消息和发送消息
	for {
		glog.Info("Receiving...")
		pack, err := packet.GetPacketFromBuffer(conn)
		if err != nil {
			glog.Errorf("Packet buffer: %v", err)
			return
		}

		tag = pack.Tag+1
		glog.Infof("Type: %v", pack.Type)
		switch pack.Type{
		case uint32(common.MessageCommand_HEARTBEAT_RESPONSE):
			heartbeatResponse := &common.HeartbeatResponse{}
			packet.Unpack(pack, heartbeatResponse)
			glog.Infof("HeartbeatResponse: %v", heartbeatResponse)
			heartbeatDelay = heartbeatResponse.GetNextHeartbeat()
		case uint32(common.MessageCommand_MESSAGE_RESPONSE):
			messageResponse := &common.MessageResponse{}
			packet.Unpack(pack, messageResponse)
			glog.Infof("MessageResponse: %v", messageResponse)
		case uint32(common.MessageCommand_RECEIVE_MESSAGE):
			message := &common.Message{}
			packet.Unpack(pack, message)
			glog.Infof("ReceiveMessage: %v", message)

			glog.Info("ReceiveMessageAck...")
			receiveMessageAck := &common.ReceiveMessageAck{
				MessageId: proto.String(message.GetMessageId()),
				Status: proto.Int32(1),
			}
			glog.Infof("ReceiveMessageAck: %v string: %s", receiveMessageAck, receiveMessageAck.String())
			err := handlers.SendPbData(conn, pack.Tag+1, uint32(common.MessageCommand_RECEIVE_MESSAGE_ACK), receiveMessageAck)
			if err != nil {
				glog.Errorf("Send data error: %v", err)
				return
			}
			glog.Info("ReceiveMessageAck")

		case uint32(common.MessageCommand_NORMARL_MESSAGE):
			normalMessage := &common.NormalMessage{}
			packet.Unpack(pack, normalMessage)
			glog.Infof("NormalMessage: %v", normalMessage)

			glog.Info("NormalMessageAck...")
			normalMessageAck := &common.NormalMessageAck{
				MessageId: proto.String(normalMessage.GetMessageId()),
				Status: proto.Int32(1),
			}
			glog.Infof("NormalMessageAck: %v string: %s", normalMessageAck, normalMessageAck.String())
			err := handlers.SendPbData(conn, pack.Tag+1, uint32(common.MessageCommand_NORMARL_MESSAGE_ACK), normalMessageAck)
			if err != nil {
				glog.Errorf("Send data error: %v", err)
				return
			}
			glog.Info("NormalMessageAck")
		case uint32(common.MessageCommand_USER_LOGOUT_RESPONSE):
			userLogoutResponse := &common.UserLogoutResponse{}
			packet.Unpack(pack, userLogoutResponse)
			glog.Infof("UserLogoutResponse: %v", userLogoutResponse)
			os.Exit(1)
		}

//		if tag > 10 {
//			userLogout(conn, sender)
//		}
	}
}

func main() {
	flag.Parse()
	glog.Info("Client start...")
	glog.Flush()
	// for i := 0; i < 1; i++ {
	// 	time.Sleep(50 * time.Millisecond)
	// 	go testBB(i)
	// }
	fmt.Println("my name is ", sender)
	fmt.Println("your name is ", receiver)
	go testBB(sender)
	fmt.Println("sleep...")
	time.Sleep(360000 * time.Second)
	glog.Info("Client end.")
}

func init() {
	flag.StringVar(&sender, "sender", "1", "自己的uuid")
	flag.StringVar(&receiver, "receiver", "2", "对方的uuid")
	flag.Parse()
	// 读取配置文件
	err := config.ReadIniFile("config.ini")
	if err != nil {
		glog.Fatalf("Error: %v", err)
	}
}

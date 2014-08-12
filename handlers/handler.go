package handlers

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/gysan/chatroom/config"
	"github.com/gysan/chatroom/dao"
	"github.com/gysan/chatroom/packet"
	"github.com/gysan/chatroom/utils/safemap"
	"net"
	"time"
	"github.com/golang/glog"
	"github.com/gysan/chatroom/common"
	"strconv"
)

var (
	// 映射conn到user, user --> conn
	UserMapConn *safemap.SafeMap = safemap.NewSafeMap()
	// 映射user到conn, conn --> user
	ConnMapUser *safemap.SafeMap = safemap.NewSafeMap()
	// ConnMapLoginStatus 映射conn的登陆状态,conn->loginstatus
	// loginstatus为nil表示conn已经登陆, 否则 loginstatus表示conn连接服务器时的时间
	// 用于判断登陆是否超时，控制恶意连接
	ConnMapLoginStatus *safemap.SafeMap = safemap.NewSafeMap()
)

//关闭conn
func CloseConn(conn *net.TCPConn) {
	glog.Info("Close connection...")
	conn.Close()
	ConnMapLoginStatus.Delete(conn)
	glog.Infof("Conn login map size: %v, Values: %v", ConnMapLoginStatus.Size(), ConnMapLoginStatus.Items())
	userId := ConnMapUser.Get(conn)
	UserMapConn.Delete(userId)
	glog.Infof("Device conn map size: %v, Values: %v", UserMapConn.Size(), UserMapConn.Items())
	ConnMapUser.Delete(conn)
	glog.Infof("Conn device map size: %v, Values: %v", ConnMapUser.Size(), ConnMapUser.Items())
	if userId != nil {
		dao.UuidOffLine(userId.(string))
		userIdInt, _ := strconv.Atoi(userId.(string))
		dao.UpdateOnline(userIdInt, 0)
	}
	glog.Info("Close connection.")
}

//初始化conn
func InitConn(conn *net.TCPConn, userId string) {
	glog.Info("Init connection...")
	ConnMapLoginStatus.Set(conn, nil)
	glog.Infof("Conn login map size: %v, Values: %v", ConnMapLoginStatus.Size(), ConnMapLoginStatus.Items())
	UserMapConn.Set(userId, conn)
	glog.Infof("Device conn map size: %v, Values: %v", UserMapConn.Size(), UserMapConn.Items())
	ConnMapUser.Set(conn, userId)
	glog.Infof("Conn device map size: %v, Values: %v", ConnMapUser.Size(), ConnMapUser.Items())
	userIdInt, _ := strconv.Atoi(userId)
	dao.UuidOnLine(userId)
	dao.UpdateOnline(userIdInt, 1)
	glog.Info("Init connection.")
}

// 发送字节流
func SendByteStream(conn *net.TCPConn, buf []byte) error {
	conn.SetWriteDeadline(time.Now().Add(time.Duration(config.WriteTimeout) * time.Second))
	n, err := conn.Write(buf)
	if n != len(buf) || err != nil {
		return fmt.Errorf("Write to %v failed, Error: %v", ConnMapUser.Get(conn).(string), err)
	}
	return nil
}

// 发送protobuf结构数据
func SendPbData(conn *net.TCPConn, tag uint32, dataType uint32, pb interface{}) error {
	pac, err := packet.Pack(tag, dataType, pb)
	if err != nil {
		return err
	}
	return SendByteStream(conn, pac.GetBytes())
}

/**
 * 心跳初始化　
 */
func HandleHeartbeatInit(conn *net.TCPConn, receivePacket *packet.Packet) {
	glog.Info("HandleHeartbeatInit beginning...")
	heartbeatInit := &common.HeartbeatInit{}
	packet.Unpack(receivePacket, heartbeatInit)
	glog.Infof("HeartbeatInit: %v string: %s", receivePacket, heartbeatInit.String())
	glog.Info("HandleHeartbeatInit end.")

	glog.Info("HeartbeatInitResponse...")
	heartbeatInitResponse := &common.HeartbeatInitResponse{
		NextHeartbeat: proto.Int32(30),
	}
	glog.Infof("HeartbeatInitResponse: %v string: %s", heartbeatInitResponse, heartbeatInitResponse.String())
	err := SendPbData(conn, receivePacket.Tag+1, uint32(common.MessageCommand_HEARTBEAT_INIT_RESPONSE), heartbeatInitResponse)
	if err != nil {
		glog.Errorf("Send data error: %v", err)
		return
	}
	glog.Info("HeartbeatInitResponse")
}

/**
 * 心跳
 */
func HandleHeartbeat(conn *net.TCPConn, receivePacket *packet.Packet) {
	glog.Info("HandleHeartbeat beginning...")
	heartbeat := &common.Heartbeat{}
	packet.Unpack(receivePacket, heartbeat)
	glog.Infof("Heartbeat: %v string: %s", receivePacket, heartbeat.String())
	glog.Info("HandleHeartbeat end.")

	conn.SetReadDeadline(time.Now().Add(time.Duration(heartbeat.GetLastDelay() + 5) * time.Second))

	glog.Info("HeartbeatResponse...")
	heartbeatResponse := &common.HeartbeatResponse{
		NextHeartbeat: proto.Int32(heartbeat.GetLastDelay() + 1),
	}
	glog.Infof("HeartbeatResponse: %v string: %s", heartbeatResponse, heartbeatResponse.String())
	err := SendPbData(conn, receivePacket.Tag+1, uint32(common.MessageCommand_HEARTBEAT_RESPONSE), heartbeatResponse)
	if err != nil {
		glog.Errorf("Send data error: %v", err)
		return
	}
	glog.Info("HeartbeatResponse end.")
}

func HandleUserLogin(conn *net.TCPConn, receivePacket *packet.Packet) {
	glog.Info("HandleUserLogin beginning...")
	userLogin := &common.UserLogin{}
	packet.Unpack(receivePacket, userLogin)
	userId := userLogin.GetUserId()

	// 检测userId合法性
	if dao.UuidCheckExist(userId) {
		// 如果已经在线则关闭以前的conn
		if dao.UuidCheckOnline(userId) {
			co := UserMapConn.Get(userId)
			if co != nil {
				CloseConn(co.(*net.TCPConn))
			}
		}
		// 上线conn
		InitConn(conn, userId)
	} else {
		CloseConn(conn)
	}

	glog.Infof("UserLogin: %v string: %s", receivePacket, userLogin.String())
	glog.Info("HandleUserLogin end.")

	glog.Info("UserLoginResponse...")
	userLoginResponse := &common.UserLoginResponse{
		Status: proto.Bool(true),
	}
	glog.Infof("UserLoginResponse: %v string: %s", userLoginResponse, userLoginResponse.String())
	err := SendPbData(conn, receivePacket.Tag+1, uint32(common.MessageCommand_USER_LOGIN_RESPONSE), userLoginResponse)
	if err != nil {
		glog.Errorf("Send data error: %v", err)
		return
	}
	glog.Info("UserLoginResponse end.")

	glog.Info("Offline message sending...")
	messages, err := dao.FindOfflineMessages(userId)
	glog.Infof("%v, %v", messages, err)

	v := &common.Message{}
	for _, v = range messages {
		pac, err := packet.Pack(receivePacket.Tag+1, uint32(common.MessageCommand_RECEIVE_MESSAGE), v)
		if err != nil {
			glog.Errorf("Packet: %v", err)
			return
		}
		SendByteStream(conn, pac.GetBytes())
	}
	glog.Info("Offline message sent.")

	glog.Info("Offline normal message sending...")
	normalMessages, err := dao.FindOfflineNormalMessages(userId)
	glog.Infof("%v, %v", normalMessages, err)

	normalMessage := &common.NormalMessage{}
	for _, normalMessage = range normalMessages {
		pac, err := packet.Pack(receivePacket.Tag+1, uint32(common.MessageCommand_NORMARL_MESSAGE), normalMessage)
		if err != nil {
			glog.Errorf("Packet: %v", err)
			return
		}
		SendByteStream(conn, pac.GetBytes())
	}
	glog.Info("Offline normal message sent.")
}

func HandleUserLogout(conn *net.TCPConn, receivePacket *packet.Packet) {
	glog.Info("HandleUserLogout beginning...")
	userLogout := &common.UserLogout{}
	packet.Unpack(receivePacket, userLogout)
	CloseConn(conn)
	glog.Info("HandleUserLogout end.")

	glog.Info("UserLogoutResponse...")
	userLogoutResponse := &common.UserLogoutResponse{
		Status: proto.Bool(true),
	}
	glog.Infof("UserLogoutResponse: %v string: %s", userLogoutResponse, userLogoutResponse.String())
	err := SendPbData(conn, receivePacket.Tag+1, uint32(common.MessageCommand_USER_LOGOUT_RESPONSE), userLogoutResponse)
	if err != nil {
		glog.Errorf("Send data error: %v", err)
		return
	}
	glog.Info("UserLogoutResponse end.")
}

func HandleMessage(conn *net.TCPConn, receivePacket *packet.Packet) {
	glog.Info("HandleMessage beginning...")
	message := &common.Message{}
	packet.Unpack(receivePacket, message)
	dao.InsertMessage(message)
	glog.Infof("Message: %v string: %s", receivePacket, message.String())
	glog.Info("HandleMessage end.")

	glog.Info("MessageResponse...")
	messageResponse := &common.MessageResponse{
		MessageId: proto.String(message.GetMessageId()),
		Status: proto.Bool(true),
	}
	glog.Infof("MessageResponse: %v string: %s", messageResponse, messageResponse.String())
	err := SendPbData(conn, receivePacket.Tag+1, uint32(common.MessageCommand_MESSAGE_RESPONSE), messageResponse)
	if err != nil {
		glog.Errorf("Send data error: %v", err)
		return
	}
	glog.Info("MessageResponse end.")

	receiver := message.GetReceiver()
	glog.Infof("Send message to receiver: %v", receiver)


	pac, err := packet.Pack(receivePacket.Tag+2, uint32(common.MessageCommand_RECEIVE_MESSAGE), message)
	if err != nil {
		glog.Errorf("Packet: %v", err)
		return
	}

	if dao.UuidCheckOnline(receiver) {
		glog.Infof("%v is online", receiver)
		receiverConn := UserMapConn.Get(receiver).(*net.TCPConn)
		if SendByteStream(receiverConn, pac.GetBytes()) != nil {
			glog.Infof("Add to offline list")
			dao.OfflineMsgAddMsg(receiver, string(pac.GetBytes()))
		}
	} else {
		glog.Infof("%v is offline", receiver)
		dao.OfflineMsgAddMsg(receiver, string(pac.GetBytes()))
	}
	glog.Info("Send message to receiver end.")
}

func HandleReceiveMessageAck(conn *net.TCPConn, receivePacket *packet.Packet) {
	glog.Info("HandleReceiveMessageAck beginning...")
	receiveMessageAck := &common.ReceiveMessageAck{}
	packet.Unpack(receivePacket, receiveMessageAck)
	dao.UpdateMessageStatus(ConnMapUser.Get(conn), receiveMessageAck.GetMessageId(), int(receiveMessageAck.GetStatus()))
	glog.Info("HandleReceiveMessageAck end.")
}

func HandleNormalMessageAck(conn *net.TCPConn, receivePacket *packet.Packet) {
	glog.Info("HandleNormalMessageAck beginning...")
	normalMessageAck := &common.NormalMessageAck{}
	packet.Unpack(receivePacket, normalMessageAck)
	dao.UpdateNormalMessageStatus(ConnMapUser.Get(conn), normalMessageAck.GetMessageId(), int(normalMessageAck.GetStatus()))
	glog.Info("HandleNormalMessageAck end.")
}

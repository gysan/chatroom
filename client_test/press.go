package main

import (
	proto "code.google.com/p/goprotobuf/proto"
	"flag"
	"fmt"
	"github.com/gysan/chatroom/config"
	"github.com/gysan/chatroom/handlers"
	"github.com/gysan/chatroom/packet"
	"github.com/gysan/chatroom/pb"
	"github.com/gysan/chatroom/utils/convert"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	prefix string // uuid的前缀为随机生成，方便多个客户端测试，避免将其他客户端踢下线
	total  int    // uuid数量, uuid名字为 1---total
)

func getUuid(uuid int) string {
	return prefix + strconv.Itoa(uuid)
}

// 发送心跳包
func ping(conn *net.TCPConn) {
	ticker := time.NewTicker(60 * time.Second)
	for _ = range ticker.C {
		//write
		writePingMsg := &pb.PbClientPing{
			Ping:      proto.Bool(true),
			Timestamp: proto.Int64(time.Now().Unix()),
		}
		err := handlers.SendPbData(conn, packet.PK_ClientPing, writePingMsg)
		if err != nil {
			return
		}
		fmt.Println(conn.RemoteAddr().String(), "ping.")
	}
}

// 处理收发数据包
func handlePackets(uuid int, conn *net.TCPConn, receivePackets <-chan *packet.Packet, chStop <-chan bool) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("Panic: %v\r\n")
		}
	}()
	for {
		select {
		case <-chStop:
			return

		// 消息包处理
		case p := <-receivePackets:
			if p.Type == packet.PK_ServerAcceptLogin { // 登陆回复
				// read
				readMsg := &pb.PbServerAcceptLogin{}
				packet.Unpack(p, readMsg)
				if readMsg.GetLogin() == true {
					log.Printf("[%v]: [%v]---[%v]\r\n", getUuid(uuid), readMsg.GetTipsMsg(), convert.TimestampToTimeString(readMsg.GetTimestamp()))
				}

				// write，随机向10个人发送消息
				for i := 0; i < 10; i++ {
					rand.Seed(time.Now().UnixNano())
					to_uuid := rand.Intn(total) + 1 // [1, total]
					if to_uuid == uuid {
						continue
					}

					writeMsg := &pb.PbC2CTextChat{
						FromUuid:  proto.String(getUuid(uuid)),
						ToUuid:    proto.String(getUuid(to_uuid)),
						TextMsg:   proto.String("hello,世界！！！"),
						Timestamp: proto.Int64(time.Now().Unix()),
					}
					handlers.SendPbData(conn, packet.PK_C2CTextChat, writeMsg)
				}

			} else if p.Type == packet.PK_C2CTextChat { // 普通消息
				// read
				readMsg := &pb.PbC2CTextChat{}
				packet.Unpack(p, readMsg)
				from_uuid := readMsg.GetFromUuid()
				to_uuid := readMsg.GetToUuid()
				txt_msg := readMsg.GetTextMsg()
				timestamp := readMsg.GetTimestamp()
				if to_uuid != getUuid(uuid) {
					log.Printf("[%v]收到[%v]发来的不属于自己的包,该包应该属于[%v]\r\n", getUuid(uuid), from_uuid, to_uuid)
				} else {
					log.Printf("[%v]：[%v]收到来自[%v]的消息: [%v]", convert.TimestampToTimeString(timestamp), getUuid(uuid), from_uuid, txt_msg)

					// write, 回复时在原基础上加点消息，控制长度范围
					var to_txt_msg string
					var add_txt string = " 你好 hello world"

					if len(txt_msg)+len(add_txt) <= 2048 {
						to_txt_msg = txt_msg + add_txt
					} else {
						to_txt_msg = txt_msg
					}

					writeMsg := &pb.PbC2CTextChat{
						FromUuid:  proto.String(getUuid(uuid)),
						ToUuid:    proto.String(from_uuid),
						TextMsg:   proto.String(to_txt_msg),
						Timestamp: proto.Int64(time.Now().Unix()),
					}
					handlers.SendPbData(conn, packet.PK_C2CTextChat, writeMsg)
				}

			} else {
				log.Printf("[%v]收到未知包\r\n", getUuid(uuid))
			}
		}
	}
}

// 模拟客户端(uuid)
func testClient(uuid int) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("uuid: [%v] Panic: %v\r\n", getUuid(uuid), e)
		}
	}()

	// 连接服务器
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", config.Addr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("[%v] DialTCP失败: %v\r\n", getUuid(uuid), err)
		return
	}

	// 发送登陆请求
	writeLoginMsg := &pb.PbClientLogin{
		Uuid:      proto.String(getUuid(uuid)),
		Version:   proto.Float32(3.14),
		Timestamp: proto.Int64(time.Now().Unix()),
	}
	err = handlers.SendPbData(conn, packet.PK_ClientLogin, writeLoginMsg)
	if err != nil {
		log.Printf("[%v] 发送登陆包失败: %v\r\n", getUuid(uuid), err)
		return
	}

	// 下面这些处理和server.go中的一样
	receivePackets := make(chan *packet.Packet, 20) // 接收到的包
	chStop := make(chan bool)                       // 通知停止消息处理
	request := make([]byte, 1024)
	buf := make([]byte, 0)
	var bufLen uint32 = 0

	defer func() {
		conn.Close()
		chStop <- true
	}()

	// 发送心跳包
	go ping(conn)

	// 处理接受到的包
	go handlePackets(uuid, conn, receivePackets, chStop)

	for {
		readSize, err := conn.Read(request)
		if err != nil {
			return
		}
		if readSize > 0 {
			buf = append(buf, request[:readSize]...)
			bufLen += uint32(readSize)

			// 包长(4) + 类型(4) + 包体(len([]byte))
			if bufLen >= 8 {
				pacLen := convert.BytesToUint32(buf[0:4])
				if bufLen >= pacLen {
					receivePackets <- &packet.Packet{
						Len:  pacLen,
						Type: convert.BytesToUint32(buf[4:8]),
						Data: buf[8:pacLen],
					}
					bufLen -= pacLen
					buf = buf[:bufLen]
				}
			}

		}
	}
}

func main() {
	for i := 1; i <= total; i++ {
		time.Sleep(50 * time.Millisecond)
		go testClient(i)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal: %v\r\n", <-ch)
}

func init() {
	flag.IntVar(&total, "t", 100, "uuid的总数")
	flag.Parse()
	// 读取配置文件
	err := config.ReadIniFile("../config.ini")
	if err != nil {
		log.Fatal(err, "\r\n")
	}

	//
	rand.Seed(time.Now().UnixNano())
	x := rand.Intn(10000)
	y := rand.Intn(10000)
	prefix = "[" + strconv.Itoa(x) + strconv.Itoa(y) + "]--> "
}

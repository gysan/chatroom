package packet

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	. "github.com/gysan/chatroom/utils/convert"
	"github.com/golang/glog"
	"errors"
	"net"
)

// ---------------------------数据包构造------------------------------------
// （1）字节序：大端模式

// （2）数据包组成：包长 + 类型 + 包体

//  包长：4字节，uint32，整个数据包的长度

//  类型：4字节，uint32

//  包体：字节数组，[]byte

//  包长和类型用明文传输，包体由结构体采用protobuf序列化后再进行AES加密得到。
// --------------------------------------------------------------------------

// 数据包的类型
const (
	PK_ClientLogin       = uint32(11)
	PK_ServerAcceptLogin = uint32(12)
	PK_ClientLogout      = uint32(13)
	PK_ClientPing        = uint32(14)
	PK_C2CTextChat       = uint32(15)
)

type Packet struct {
	Tag     uint32
	Type    uint32
	Length  uint32
	Data    []byte
}

// 得到序列化后的Packet
func (this *Packet) GetBytes() (buf []byte) {
	buf = append(buf, Uint32ToBytes(this.Tag)...)
	buf = append(buf, Uint32ToBytes(this.Type)...)
	buf = append(buf, Uint32ToBytes(this.Length)...)
	buf = append(buf, this.Data...)
	glog.Infof("Buffer: %v", buf)
	return buf
}

// 将数据包类型和pb数据结构一起打包成Packet，并加密Packet.Data
func Pack(tag uint32, dataType uint32, pb interface{}) (*Packet, error) {
	pbData, err := proto.Marshal(pb.(proto.Message))
	if err != nil {
		return nil, fmt.Errorf("Pack error: %v", err.Error())
	}

	pac := &Packet{}
	pac.Tag = tag
	pac.Type = dataType
	// 对Data进行AES加密
	//	pac.Data, err = aes.AesEncrypt(pbData)
	//	if err != nil {
	//		return nil, fmt.Errorf("Pack error: %v", err.Error())
	//	}
	pac.Data = pbData
	pac.Length = uint32(len(pac.Data))
	glog.Infof("Pack: %v", pac)
	return pac, nil
}

// 将Packet解包成非加密的pb数据结构
func Unpack(pac *Packet, pb interface{}) error {
	if pac == nil {
		return fmt.Errorf("Unpack error: pac == nil")
	}

	//	decryptData, err := aes.AesDecrypt(pac.Data)
	//	if err != nil {
	//		return fmt.Errorf("Unpack error: %v", err.Error())
	//	}

	err := proto.Unmarshal(pac.Data, pb.(proto.Message))
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}
	return nil
}

func GetPacketFromBuffer(conn *net.TCPConn) (*Packet, error) {
	head := make([]byte, 12)
	conn.Read(head)
	tag := BytesToUint32(head[0:4])
	if tag == 0 {
		return nil, errors.New("Buffer not exist")
	}

	packetLength := BytesToUint32(head[8:12])
	body := make([]byte, packetLength)
	conn.Read(body)
	pack := &Packet{
		Tag: BytesToUint32(head[0:4]),
		Type: BytesToUint32(head[4:8]),
		Length:  packetLength,
		Data: body,
	}
	glog.Infof("Tag: %v, Type: %v, Length: %v, Data: %v, Data string: %s", pack.Tag, pack.Type, pack.Length, pack.Data, pack.Data)
	return pack, nil
}

package packet

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/gysan/chatroom/pb"
	"github.com/gysan/chatroom/utils/convert"
	"testing"
	"time"
)

func TestPacket(t *testing.T) {
	pbData := &pb.PbServerAcceptLogin{
		Login:     proto.Bool(true),
		TipsMsg:   proto.String("hello世界 哈哈，我爱你"),
		Timestamp: proto.Int64(time.Now().Unix()),
	}

	pac, err := Pack(PK_ServerAcceptLogin, pbData)
	if err != nil {
		t.Error(err)
	}

	ppd := &pb.PbServerAcceptLogin{}
	Unpack(pac, ppd)

	fmt.Println("Login:", ppd.GetLogin())
	fmt.Println("TipsMsg:", ppd.GetTipsMsg())
	fmt.Println("Timestamp:", convert.TimestampToTimeString(ppd.GetTimestamp()))
}

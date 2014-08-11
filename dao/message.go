package dao

import (
	"database/sql"
	"github.com/golang/glog"
	"github.com/gysan/chatroom/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gysan/chatroom/common"
	"code.google.com/p/goprotobuf/proto"
	"time"
	"github.com/gysan/chatroom/utils/convert"
	"strconv"
)

var messageDb *sql.DB

func GetMessageDb() *sql.DB {
	var err error
	if messageDb == nil {
		messageDb, err = sql.Open("mysql", config.Source)
		if err != nil {
			glog.Errorf("sql.Open(\"mysql\", %s) failed (%v)", config.Source, err)
			panic(err)
		}
		glog.Infof("sql.Open(\"mysql\", %s)", config.Source)
	}
	return messageDb
}

func FindOfflineMessages(userId string) ([]*common.Message, error) {
	userIdInt, _ := strconv.Atoi(userId)
	glog.Infof("Find offline messages: %v", userIdInt%16)
	rows, err := GetMessageDb().Query("select id, sender, receiver, body from marry_me_message_"+strconv.Itoa(userIdInt%16)+" where sender = ? and status = 0 order by ctime asc", userId)
	if err != nil {
		glog.Errorf("db.Query(\"%s\") failed (%v)", "select id, sender, receiver, body from marry_me_message_"+strconv.Itoa(userIdInt%16)+" where sender = ? and status = 0 order by ctime asc", err)
		return nil, err
	}
	var id, sender, receiver, body string
	messages := []*common.Message{}
	for rows.Next() {
		if err := rows.Scan(&id, &receiver, &sender, &body); err != nil {
			glog.Errorf("rows.Scan() failed (%v)", err)
			return nil, err
		}
		glog.Infof("%v", id)
		messages = append(messages, &common.Message{MessageId: proto.String(id), Sender: proto.String(sender), Receiver: proto.String(receiver), MessageBody: proto.String(body)})
	}
	return messages, nil
}

func UpdateStatus(receiver interface {}, messageId string, status int) error {
	receiverInt, _ := strconv.Atoi(receiver.(string))
	glog.Infof("Update status: %v, %v", receiverInt%16, status)
	timestamp := convert.TimestampToTimeString(time.Now().Unix())
	_, err := GetMessageDb().Exec("update marry_me_message_"+strconv.Itoa(receiverInt%16)+" set status = ?, utime = ? where id = ? and sender = ?", status, timestamp, messageId, receiver.(string))
	if err != nil {
		glog.Errorf("db.Exec(\"update marry_me_message_"+strconv.Itoa(receiverInt%16)+" set status = ?, utime = ? where id = ? and sender = ?\") failed (%v)", err)
		return err
	}
	return nil
}

func InsertMessage(message *common.Message) error {
	senderInt, _ := strconv.Atoi(message.GetSender())
	receiverInt, _ := strconv.Atoi(message.GetReceiver())
	glog.Infof("Insert message: [sender: %v] [receiver: %v]", senderInt%16, receiverInt%16)
	timestamp := convert.TimestampToTimeString(time.Now().Unix())
	_, err := GetMessageDb().Exec("INSERT INTO `marry_me_message_"+strconv.Itoa(senderInt%16)+"` (`id`, `sender`, `receiver`, `body`, `status`, `utime`, `ctime`) VALUES (?, ?, ?, ?, ?, ?, ?)",
		message.GetMessageId(), message.GetSender(), message.GetReceiver(), message.GetMessageBody(), 10, timestamp, timestamp)
	_, senderErr := GetMessageDb().Exec("INSERT INTO `marry_me_message_"+strconv.Itoa(receiverInt%16)+"` (`id`, `sender`, `receiver`, `body`, `status`, `utime`, `ctime`) VALUES (?, ?, ?, ?, ?, ?, ?)",
		message.GetMessageId(), message.GetReceiver(), message.GetSender(), message.GetMessageBody(), 0, timestamp, timestamp)
	if err != nil || senderErr != nil {
		glog.Errorf("db.Exec(\"INSERT INTO `marry_me_message` (`id`, `sender`, `is_sender`, `receiver`, `body`, `utime`, `ctime`) VALUES (?, ?, ?, ?, ?, ?, ?)\") failed (%v)", err)
		return err
	}
	return nil
}

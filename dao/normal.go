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

var normalMessageDb *sql.DB

func GetNormalMessageDb() *sql.DB {
	var err error
	if normalMessageDb == nil {
		normalMessageDb, err = sql.Open("mysql", config.Source)
		if err != nil {
			glog.Errorf("sql.Open(\"mysql\", %s) failed (%v)", config.Source, err)
			panic(err)
		}
		glog.Infof("sql.Open(\"mysql\", %s)", config.Source)
	}
	return normalMessageDb
}

func FindOfflineNormalMessages(userId string) ([]*common.NormalMessage, error) {
	userIdInt, _ := strconv.Atoi(userId)
	glog.Infof("Find offline messages: %v", userIdInt%16)
	rows, err := GetNormalMessageDb().Query("select id, receiver, body, ctime from marry_me_normal_message_"+strconv.Itoa(userIdInt%16)+" where receiver = ? and status = 0 order by ctime asc", userId)
	if err != nil {
		glog.Errorf("db.Query(\"%s\") failed (%v)", "select id, receiver, body, ctime from marry_me_normal_message_"+strconv.Itoa(userIdInt%16)+" where receiver = ? and status = 0 order by ctime asc", err)
		return nil, err
	}
	var id, receiver, body string
	var date time.Time
	messages := []*common.NormalMessage{}
	for rows.Next() {
		if err := rows.Scan(&id, &receiver, &body, &date); err != nil {
			glog.Errorf("rows.Scan() failed (%v)", err)
			return nil, err
		}
		glog.Infof("%v", id)
		messages = append(messages, &common.NormalMessage{MessageId: proto.String(id), Receiver: proto.String(receiver), Content: []byte(body), Date: proto.Int64(date.Unix())})
	}
	return messages, nil
}

func UpdateNormalMessageStatus(receiver interface{}, messageId string, status int) error {
	receiverInt, _ := strconv.Atoi(receiver.(string))
	glog.Infof("Update status: %v, %v", receiverInt%16, status)
	timestamp := convert.TimestampToTimeString(time.Now().Unix())
	_, err := GetNormalMessageDb().Exec("update marry_me_normal_message_"+strconv.Itoa(receiverInt%16)+" set status = ?, utime = ? where id = ?", status, timestamp, messageId)
	if err != nil {
		glog.Errorf("db.Exec(\"update marry_me_normal_message_"+strconv.Itoa(receiverInt%16)+" set status = ?, utime = ? where id = ?\") failed (%v)", err)
		return err
	}
	return nil
}

func InsertNormalMessage(message *common.NormalMessage) error {
	receiverInt, _ := strconv.Atoi(message.GetReceiver())
	glog.Infof("Insert message: [receiver: %v]", receiverInt%16)
	timestamp := convert.TimestampToTimeString(time.Now().Unix())
	_, err := GetNormalMessageDb().Exec("INSERT INTO `marry_me_normal_message_"+strconv.Itoa(receiverInt%16)+"` (`id`, `receiver`, `body`, `status`, `utime`, `ctime`) VALUES (?, ?, ?, ?, ?, ?)",
		message.GetMessageId(), message.GetReceiver(), message.GetContent(), 0, timestamp, timestamp)
	if err != nil {
		glog.Errorf("db.Exec(\"INSERT INTO `marry_me_message` (`id`, `receiver`, `body`, `utime`, `ctime`) VALUES (?, ?, ?, ?, ?, ?, ?)\") failed (%v)", err)
		return err
	}
	return nil
}

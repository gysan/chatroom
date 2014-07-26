//
// 处理消息记录(msgid, msg)
// msgid为msg的编号, msg为消息的实际内容
// 数据表：data(msgid varchar(32) primary key, msg varchar(1024))
// 只与数据库交互数据
//
package dao

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
)

var (
	id_msg_db   *sql.DB
	id_msg_lock *sync.RWMutex = new(sync.RWMutex)
)

// 初始化
func IdMsgInit(dbPath string) {
	var err error
	id_msg_db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err, "\r\n")
	}
	id_msg_db.Exec("create table if not exists data(msgid varchar(32) primary key, msg varchar(1024))")
}

// 程序结束时清理
func IdMsgClean() {
	id_msg_lock.Lock()
	defer id_msg_lock.Unlock()
	id_msg_db.Close()
}

// 增加(msgid, msg)
func IdMsgAdd(msgid, msg string) {
	id_msg_lock.Lock()
	defer id_msg_lock.Unlock()
	// 注意这里的msg中包含了各种不可见字符，各种恶心字符。
	// 下面这种做法将导致msg不能插入到数据库中
	// id_msg_db.Exec("insert into data(msgid, msg) values(?, ?)", msgid, msg)
	// 需要将 msg 由string转换成 []byte， 才能插入数据库。
	// 根据string读出msg还是没问题的。
	// 惨痛的教训，切记
	id_msg_db.Exec("insert into data(msgid, msg) values(?, ?)", msgid, []byte(msg))
}

// 增加多个(msgid, msg)
func IdMsgMultiAdd(msgids, msgs []string) {
	id_msg_lock.Lock()
	defer id_msg_lock.Unlock()
	// 启动事务来插入数据
	tx, err := id_msg_db.Begin()
	if err != nil {
		log.Printf("%v\r\n", err)
	} else {
		for i, _ := range msgids {
			tx.Exec("insert into data(msgid, msg) values(?, ?)", msgids[i], []byte(msgs[i]))
		}
		tx.Commit()
	}
}

// 删除msg
func IdMsgDelete(msgid string) {
	id_msg_lock.Lock()
	defer id_msg_lock.Unlock()
	id_msg_db.Exec("delete * from data where msgid = ?", msgid)
}

// 删除多个msg
func IdMsgMultiDelete(msgids []string) {
	id_msg_lock.Lock()
	defer id_msg_lock.Unlock()
	tx, err := id_msg_db.Begin()
	if err != nil {
		log.Printf("%v\r\n", err)
	} else {
		for i, _ := range msgids {
			tx.Exec("delete * from data where msgid = ?", msgids[i])
		}
		tx.Commit()
	}
}

// 查找(msgid, msg)
func IdMsgGetMsgFromId(msgid string) string {
	id_msg_lock.RLock()
	defer id_msg_lock.RUnlock()
	var msg string
	err := id_msg_db.QueryRow("select msg from data where msgid = ?", msgid).Scan(&msg)
	if err != nil {
		log.Printf("%v\r\n", err)
		return ""
	}
	return msg
}

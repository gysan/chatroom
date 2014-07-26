//
// 处理离线消息, 每条消息的编号都是全局唯一的
// 编号是 消息加上一些额外的信息(时间，uuid等)计算出来的md5值
// 数据表：data(msgid varchar(32) primary key, uuid varchar(32))
// uuid的离线消息编号保存在内存中， uuid --> msgid1, msgid2, msgid3, ...
// 定时，异步更新数据库
//
package dao

import (
	"database/sql"
	"github.com/gysan/chatroom/config"
	"github.com/gysan/chatroom/utils/convert"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
	"time"
)

// offline_msg_stop 程序结束的信号
// offline_msg_uuid2msgids 保存uuid的离线消息，uuid --> msgid1, msgid2, msgid3, ...
// offline_msg_deletemsgids 缓存了待删除的消息编号，定时更新数据库
// offline_msg_addedmsgids 缓存了待增加的离线信息，定时更新数据库
var (
	offline_msg_db           *sql.DB
	offline_msg_lock         *sync.RWMutex       = new(sync.RWMutex)
	offline_msg_stop         chan bool           = make(chan bool)
	offline_msg_uuid2msgids  map[string][]string = make(map[string][]string)
	offline_msg_addedmsgids  map[string][]string = make(map[string][]string)
	offline_msg_deletemsgids []string            = make([]string, 0)
)

// 初始化
func OfflineMsgInit(dbPath string) {
	var err error
	offline_msg_db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err, "\r\n")
	}
	offline_msg_db.Exec("create table if not exists data(msgid varchar(32) primary key, uuid varchar(32))")
	rows, err := offline_msg_db.Query("select * from data")
	if err != nil {
		log.Fatal(err, "\r\n")
	}
	defer rows.Close()

	var (
		msgid string
		uuid  string
	)
	for rows.Next() {
		rows.Scan(&msgid, &uuid)
		offline_msg_uuid2msgids[uuid] = append(offline_msg_uuid2msgids[uuid], msgid)
	}

	go offlineMsgTimedDo()
}

// 程序结束时清理
func OfflineMsgClean() {
	close(offline_msg_stop)
	offlineMsgWriteDB()
	offline_msg_db.Close()
}

// 定时更新数据库
func offlineMsgTimedDo() {
	ticker := time.NewTicker(time.Duration(config.TimedUpdateDB) * time.Second)
	for {
		select {
		case <-ticker.C:
			offlineMsgWriteDB()
		case <-offline_msg_stop:
			return
		}
	}
}

// 将新增的记录和删除的记录更新到数据库中
func offlineMsgWriteDB() {
	offline_msg_lock.Lock()
	defer offline_msg_lock.Unlock()

	// 新增记录
	tx, err := offline_msg_db.Begin()
	if err != nil {
		log.Printf("%v\r\n", err)
	} else {
		for uuid, msgids := range offline_msg_addedmsgids {
			for i, _ := range msgids {
				tx.Exec("insert into data(msgid, uuid) values(?, ?)", msgids[i], uuid)
			}
		}
		tx.Commit()
		for k, _ := range offline_msg_addedmsgids {
			delete(offline_msg_addedmsgids, k)
		}
	}

	// 删除记录
	tx, err = offline_msg_db.Begin()
	if err != nil {
		log.Printf("%v\r\n", err)
	} else {
		for i, _ := range offline_msg_deletemsgids {
			tx.Exec("delete * from data where msgid = ?", offline_msg_deletemsgids[i])
		}
		tx.Commit()
		IdMsgMultiDelete(offline_msg_deletemsgids)
		offline_msg_deletemsgids = nil
	}
}

// 检查uuid是否存在离线消息
func OfflineMsgCheck(uuid string) bool {
	offline_msg_lock.RLock()
	defer offline_msg_lock.RUnlock()
	_, ok := offline_msg_uuid2msgids[uuid]
	return ok
}

// 得到uuid的所有离线消息编号
func OfflineMsgGetIds(uuid string) []string {
	offline_msg_lock.RLock()
	defer offline_msg_lock.RUnlock()
	return offline_msg_uuid2msgids[uuid]
}

// 删除uuid的在指定msgid之前（含本身）的所有离线消息编号
// 保证msdid一定存在
func OfflineMsgDeleteIds(uuid, msgid string) {
	if msgid == "" {
		return
	}
	offline_msg_lock.Lock()
	defer offline_msg_lock.Unlock()

	var (
		k int = 0
		v string
	)
	for k, v = range offline_msg_uuid2msgids[uuid] {
		if v == msgid {
			break
		}
	}

	offline_msg_deletemsgids = append(offline_msg_deletemsgids, offline_msg_uuid2msgids[uuid][0:k+1]...)
	offline_msg_uuid2msgids[uuid] = offline_msg_uuid2msgids[uuid][k+1:]
	if len(offline_msg_uuid2msgids[uuid]) == 0 {
		delete(offline_msg_uuid2msgids, uuid)
	}
}

// 添加离线消息编号
func offlineMsgAdd(uuid, msgid string) {
	offline_msg_lock.Lock()
	defer offline_msg_lock.Unlock()
	offline_msg_addedmsgids[uuid] = append(offline_msg_addedmsgids[uuid], msgid)
	offline_msg_uuid2msgids[uuid] = append(offline_msg_uuid2msgids[uuid], msgid)
}

// 添加离线消息
func OfflineMsgAddMsg(uuid, msg string) {
	msgid := convert.StringToMd5(uuid + msg + time.Now().String())
	IdMsgAdd(msgid, msg)
	offlineMsgAdd(uuid, msgid)
}

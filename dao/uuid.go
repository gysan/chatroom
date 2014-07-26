//
// 与用户唯一标识符(uuid)相关的信息处理， 上线与离线的处理等
// 数据表：data(uuid varchar(32) primary key, registertime timestamp), registertime为uuid的注册时间
// 所有uuid信息保存在内存中, 更新时实时写入数据库
//
package dao

import (
	"database/sql"
	"github.com/gysan/chatroom/utils/convert"
	"github.com/gysan/chatroom/utils/safemap"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
	"time"
)

var (
	uuid_db             *sql.DB
	uuid_lock           *sync.RWMutex    = new(sync.RWMutex)
	uuid_allUsers       *safemap.SafeMap = safemap.NewSafeMap() // 所有用户的uuid信息(uuid, registertime)
	uuid_allOnlineUsers *safemap.SafeMap = safemap.NewSafeMap() // 在线用户, (uuid, true)
)

// 初始化
func UuidInit(dbPath string) {
	var err error
	uuid_db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err, "\r\n")
	}

	uuid_db.Exec("create table if not exists data(uuid varchar(32) primary key, registertime timestamp)")
	rows, err := uuid_db.Query("select * from data")
	if err != nil {
		log.Fatal(err, "\r\n")
	}
	defer rows.Close()

	var (
		uuid         string
		registertime time.Time
	)
	for rows.Next() {
		rows.Scan(&uuid, &registertime)
		uuid_allUsers.Set(uuid, registertime)
	}
}

// 程序结束时清理
func UuidClean() {
	uuid_lock.Lock()
	defer uuid_lock.Unlock()
	uuid_db.Close()
}

// 向数据库增加(uuid, registertime)
func uuidAdd(uuid string, registertime time.Time) {
	uuid_lock.Lock()
	defer uuid_lock.Unlock()
	uuid_db.Exec("insert into data(uuid, registertime) values(?, ?)", uuid, registertime)
}

// 向web服务器请求验证uuid是否存在, 并返回uuid的注册时间戳
func uuidCheckExistFromWeb(uuid string) (bool, int64) {
	// 发送get请求
	// ...
	return true, time.Now().Unix()
}

// 判断uuid是否存在
func UuidCheckExist(uuid string) bool {
	if uuid_allUsers.Check(uuid) {
		return true
	}
	if ok, timestamp := uuidCheckExistFromWeb(uuid); ok {
		registertime := convert.TimestampToTime(timestamp)
		uuidAdd(uuid, registertime)
		return true
	}
	return false
}

// 判断uuid是否在线
func UuidCheckOnline(uuid string) bool {
	return uuid_allOnlineUsers.Check(uuid)
}

// uuid上线
func UuidOnLine(uuid string) {
	uuid_allOnlineUsers.Set(uuid, true)
}

// uuid下线
func UuidOffLine(uuid string) {
	uuid_allOnlineUsers.Delete(uuid)
}

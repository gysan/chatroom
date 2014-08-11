package dao

import (
	"github.com/gysan/chatroom/config"
	"github.com/golang/glog"
	"database/sql"
)

var userDb *sql.DB

func GetUserDb() *sql.DB {
	var err error
	if userDb == nil {
		userDb, err = sql.Open("mysql", config.Source)
		if err != nil {
			glog.Errorf("sql.Open(\"mysql\", %s) failed (%v)", config.Source, err)
			panic(err)
		}
		glog.Infof("sql.Open(\"mysql\", %s)", config.Source)
	}
	return userDb
}

func UpdateOnline(userId, status int) error {
	glog.Infof("Update status: %v, %v", userId, status)
	_, err := GetUserDb().Exec("update marry_me_user set online = ? where id = ?", status, userId)
	if err != nil {
		glog.Errorf("db.Exec(\"update marry_me_user set online = ? where id = ?\") failed (%v)", err)
		return err
	}
	return nil
}

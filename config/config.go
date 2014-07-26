package config

import (
	"fmt"
	"github.com/robfig/config"
)

var (
	Addr            string
	Http            string
	AcceptTimeout   int
	ReadTimeout     int
	WriteTimeout    int

	Source          string

	UuidDB          string
	OfflineMsgidsDB string
	IdToMsgDB       string
	TimedUpdateDB   int
	LogFile         string
)

func ReadIniFile(inifile string) error {
	conf, err := config.ReadDefault(inifile)
	if err != nil {
		return fmt.Errorf("Read %v error. %v", inifile, err.Error())
	}

	Addr, _ = conf.String("service", "addr")
	Http, _ = conf.String("service", "http")
	AcceptTimeout, _ = conf.Int("service", "accept_timeout")
	ReadTimeout, _ = conf.Int("service", "read_timeout")
	WriteTimeout, _ = conf.Int("service", "write_timeout")

	Source, _ = conf.String("mysql", "source")

	UuidDB, _ = conf.String("data", "uuid_db")
	OfflineMsgidsDB, _ = conf.String("data", "offline_msgids_db")
	IdToMsgDB, _ = conf.String("data", "id_to_msg_db")
	TimedUpdateDB, _ = conf.Int("data", "timed_update_db")

	LogFile, _ = conf.String("debug", "logfile")

	return nil
}

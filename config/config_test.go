package config

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	err := ReadIniFile("../config.ini")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%q\n", Addr)
	fmt.Printf("%v\n", NumCpu)
	fmt.Printf("%v\n", AcceptTimeout)
	fmt.Printf("%v\n", ReadTimeout)
	fmt.Printf("%v\n", WriteTimeout)
	fmt.Printf("%q\n", UuidDB)
	fmt.Printf("%q\n", OfflineMsgidsDB)
	fmt.Printf("%q\n", IdToMsgDB)
	fmt.Printf("%v\n", TimedUpdateDB)
	fmt.Printf("%q\n", LogFile)
}

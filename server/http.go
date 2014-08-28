package server

import (
	"github.com/golang/glog"
	"net/http"
	"net"
	"encoding/json"
	"time"
	"io/ioutil"
	"github.com/gysan/chatroom/config"
	"github.com/gysan/chatroom/common"
	"github.com/gysan/chatroom/packet"
	"code.google.com/p/goprotobuf/proto"
	"github.com/gysan/chatroom/handlers"
	"strconv"
	"github.com/gysan/chatroom/dao"
	"github.com/gysan/chatroom/utils/convert"
)

func StartHttp() {
	glog.Info("Http start...")
	pushServeMux := http.NewServeMux()
	pushServeMux.HandleFunc("/push", Push)
	go httpListen(pushServeMux, config.Http)
}

func Push(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	glog.Infof("%v", params)
	receiver := params.Get("receiver")
	date := params.Get("date")
	res := map[string]interface{}{"result": 0}
	defer returnWrite(w, r, res)
	if receiver == "" {
		res["result"] = -1
		return
	}
	body := ""
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	body = string(bodyBytes)
	if body == "" {
		res["result"] = -2
		return
	}
	glog.Infof("%v", body)

	i, err := strconv.ParseInt(date, 10, 64)
	if err != nil {
		panic(err)
	}

	messageId := convert.StringToMd5(receiver+time.Now().String())

	normalMessage := &common.NormalMessage{
		MessageId: proto.String(messageId),
		Receiver: proto.String(receiver),
		Content: []byte(body),
		Date: proto.Int64(i),
	}
	pac, err := packet.Pack(uint32(1), uint32(common.MessageCommand_NORMARL_MESSAGE), normalMessage)
	if err != nil {
		glog.Errorf("Packet: %v", err)
		return
	}

	dao.InsertNormalMessage(normalMessage)

	receiverConn := handlers.UserMapConn.Get(receiver).(*net.TCPConn)
	handlers.SendByteStream(receiverConn, pac.GetBytes())
}

func httpListen(mux *http.ServeMux, bind string) {
	glog.Info("Http listen start:")
	server := &http.Server{Handler: mux, ReadTimeout: 30 * time.Second}
	l, err := net.Listen("tcp", bind)
	glog.Infof("Http start listen on %v", l.Addr())
	if err != nil {
		glog.Errorf("net.Listen('tcp', '%s') error(%v)", bind, err)
		panic(err)
	}
	if err := server.Serve(l); err != nil {
		glog.Errorf("server.Serve() error(%v)", err)
		panic(err)
	}
	glog.Info("Http listen end.")
}

func returnWrite(responseWrite http.ResponseWriter, request *http.Request, result map[string]interface{}) {
	data, err := json.Marshal(result)
	if err != nil {
		glog.Errorf("json.Marshal(\"%v\") error(%v)", result, err)
		return
	}

	if n, err := responseWrite.Write([]byte(string(data))); err != nil {
		glog.Errorf("w.Write(\"%s\") error(%v)", string(data), err)
	} else {
		glog.V(1).Infof("w.Write(\"%s\") write %d bytes", string(data), n)
	}
	glog.Infof("url: \"%s\", result:\"%s\", ip:\"%s\"", request.URL.String(), string(data), request.RemoteAddr)
}

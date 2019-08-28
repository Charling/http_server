//游戏http服务，用于接收web后台充值之类的消息处理
package http_server

import (
	"io/ioutil"
	"log"
	"net/http"

	VP "github.yn.com/ext/common/proto"

	"os"
	"path"

	BASE "github.yn.com/ext/common/function"
	LOGGER "github.yn.com/ext/common/logger"

	"fmt"

	"github.com/golang/protobuf/proto"
	"github.yn.com/ext/common/gomsg"
)

var (
	mapObjects map[int32]*Object

	Logger *log.Logger
	_file  *os.File
)

func (obj *Object) Startups(httpUrl string, msgHand *map[string]Handler, msgHead *map[int32]MsgHandler) {
	obj.handlers = *msgHand
	obj.handlers["/"] = onDispatch
	if msgHead != nil {
		obj.msgHandlers = *msgHead
	}

	go acceptHTTPRequest(obj, httpUrl)
}

// acceptHTTPRequest 监听和接受HTTP
func acceptHTTPRequest(obj *Object, httpUrl string) {
	defer gomsg.Recover()
	s := &http.Server{
		Addr:    httpUrl,
		Handler: &httpHandler{
			ID: obj.ID,
		},
		// ReadTimeout:    10 * time.Second,
		//WriteTimeout:   120 * time.Second,
		MaxHeaderBytes: 1 << 8,
	}

	LOGGER.Info("Http server listen at:%s,id:%d\n", httpUrl, obj.ID)
	err := s.ListenAndServe()
	if err != nil {
		LOGGER.Error("Http server ListenAndServe %s failed:%v\n", httpUrl, err)
	}
}

// ServeHTTP HTTP 请求处理
func (hh *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer gomsg.Recover()

	//允许http跨域访问
	w.Header().Set("Access-Control-Allow-Origin", "*") 

	ip := GetIP(r)
	if obj,ok := mapObjects[hh.ID]; ok {
		arrlen := len(obj.ips)
		find := true
		for i:=0; i < arrlen; i++ {
			if obj.ips[i] == ip {
				find = false
			}
		}
		if find == false {
			LOGGER.Warning("ip:%s invalid.", ip)
			sendResponse(w, 404, []byte(fmt.Sprintf(`{"res": "ip(%s) invalid."}`, ip)))
			return
		}
			
		var requestPath = r.URL.Path
		log.Printf(requestPath)
		requestPath = path.Base(requestPath)
		handler := obj.handlers[requestPath]
		LOGGER.Info("requestPath:%s.", requestPath)

		if handler != nil {
			handler(obj, w, r)
		} else {
			sendResponse(w, 404, []byte(fmt.Sprintf(`{"res": "oh shit, no handler."}`)))
			LOGGER.Error("oh shit, no handler.")
		}
	}
}

func onDispatch(obj *Object, w http.ResponseWriter, r *http.Request) {
	defer gomsg.Recover()
	defer r.Body.Close()
	vals := r.URL.Query()
	ops, success := BASE.Query(&vals, "ops", w)
	if !success {
		return
	}
	playerId := int64(0)
	val := vals.Get("playerid")
	if val != "" {
		playerId = BASE.StrToInt64(val)
	}
	var (
		en   int32
		data []byte
		size int32
	)

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendResponse(w, 404, []byte(fmt.Sprintf(`{"res": "ioutil.ReadAll(r.Body) != err."}`)))
		LOGGER.Error("onDispatch  res ops=%s ioutil.ReadAll(r.Body) error:%v.", ops, err)
		return
	}

	handler, exist := obj.msgHandlers[BASE.StrToInt32(ops)]
	if exist {
		en, data, size = handler(w, r, bytes, playerId)
	} else {
		en = int32(VP.ErrorCode_NoHandler)
		data = nil
		size = 0
	}

	send := &VP.HttpResult {
		En: proto.Int32(en),
		Data: data,
		Size: size,
	}
	res, err := proto.Marshal(send)
	if err != nil {
		e := fmt.Sprintf("Mashal data error %v", err)
		LOGGER.Error(e)
	}
	sendResponse(w, 202, res)
}

func sendResponse(w http.ResponseWriter, code int, data []byte) {
	w.WriteHeader(code)
	w.Write(data)
}

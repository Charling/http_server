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
	//"github.yn.com/ext/common/LOGGER"
	//"github.com/golang/protobuf/proto"
	"strings"
	token "github.yn.com/ext/common/token"
	"encoding/json"
)

var (
	mapObjects map[int32]*Object

	Logger *log.Logger
	_file  *os.File

)

func (obj *Object) Startups(httpUrl string, msgHand *map[string]Handler, msgHead *map[int32]MsgHandler) {
	obj.handlers = *msgHand
	obj.handlers["/"] = onDispatch
	obj.handlers["hall"] = onDispatch
	if msgHead != nil {
		obj.msgHandlers = *msgHead
	}

	go acceptHTTPRequest(obj, httpUrl)
}

func (obj *Object) StartJsonups(httpUrl string, msgHand *map[string]Handler, msgHead *map[int32]MsgJsonHandler) {
	obj.handlers = *msgHand
	obj.handlers["/"] = onDispatch
	obj.handlers["hall"] = onDispatch
	if msgHead != nil {
		obj.msgJsonHandlers = *msgHead
	}

	go acceptHTTPRequest(obj, httpUrl)
}

func (obj *Object) StartupGss(gameid, arenaid int32, httpUrl string, msgHand *map[string]Handler, msgHead *map[int32]MsgHandler) {
	obj.handlers = *msgHand
	obj.handlers["/"] = onDispatch
	gsUrl := fmt.Sprintf("gs_%d_%d", gameid, arenaid)
	obj.handlers[gsUrl] = onDispatch
	if msgHead != nil {
		obj.msgHandlers = *msgHead
	}

	go acceptHTTPRequest(obj, httpUrl)
}

func (obj *Object) StartupJsonGss(gameid, arenaid int32, httpUrl string, msgHand *map[string]Handler, msgHead *map[int32]MsgJsonHandler) {
	obj.handlers = *msgHand
	obj.handlers["/"] = onDispatch
	gsUrl := fmt.Sprintf("gs_%d_%d", gameid, arenaid)
	obj.handlers[gsUrl] = onDispatch
	if msgHead != nil {
		obj.msgJsonHandlers = *msgHead
	}

	go acceptHTTPRequest(obj, httpUrl)
}
// acceptHTTPRequest 监听和接受HTTP
func acceptHTTPRequest(obj *Object, httpUrl string) {
	defer LOGGER.Recover()

	s := &http.Server{
		Addr:    httpUrl,
		Handler: &httpHandler{
			ID: obj.ID,
		},
		// ReadTimeout:    10 * time.Second,
		//WriteTimeout:   120 * time.Second,
		MaxHeaderBytes: 1 << 8,
	}

/*	if obj.Type == int32(Client) {

		LOGGER.Info("Http server listen at:%s,id:%d\n", httpUrl, obj.ID)
		err := s.ListenAndServeTLS("cert/server.crt", "cert/server.key")
		if err != nil {
			LOGGER.Error("Http server ListenAndServe %s failed:%v\n", httpUrl, err)
		}

	} else {
		*/
		LOGGER.Info("Http server listen at:%s,id:%d\n", httpUrl, obj.ID)
		err := s.ListenAndServe()
		if err != nil {
			LOGGER.Error("Http server ListenAndServe %s failed:%v\n", httpUrl, err)
		}
//	}	
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer LOGGER.Recover()

	//fmt.Fprintf(w, "Hi, This is an example of https service in golang!")
	//fmt.Sprintf("这是一个https请求")
		
	//允许http跨域访问
	w.Header().Set("Access-Control-Allow-Origin", "*") 

//	ip := GetIP(r)
//	LOGGER.Info("ip:%s", ip)
	for _,obj := range mapObjects {
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

// ServeHTTP HTTP 请求处理
func (hh *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer LOGGER.Recover()

	//fmt.Fprintf(w, "Hi, This is an example of https service in golang!")
	//fmt.Sprintf("这是一个https请求")
		
	//允许http跨域访问
	w.Header().Set("Access-Control-Allow-Origin", "*") 

	ip := GetIP(r)
//	LOGGER.Info("ip:%s", ip)
	if obj,ok := mapObjects[hh.ID]; ok {
		arrlen := len(obj.ips)
		if arrlen > 0 {
			find := false
			for i:=0; i < arrlen; i++ {
				if strings.Compare(obj.ips[i],ip) == 0 {
					find = true
				}
			}
			if find == false {
			//	LOGGER.Error("ip:%s invalid.", ip)
				sendResponse(w, 404, []byte(fmt.Sprintf(`{"res": "ip(%s) invalid."}`, ip)))
				return
			}
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
	defer LOGGER.Recover()
	defer r.Body.Close()
	//vals := r.URL.Query()
	
	if r.Method == "POST" {
		dispatchPost(obj, w, r)
	} else if r.Method == "GET" {
		dispatchGet(obj, w, r)
	} else {
		dispatchPost(obj, w, r)
	}
}

func dispatchGet(obj *Object, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vals := r.URL.Query()

	startTicks := BASE.GetCurTicks()

	iplayerid, success := BASE.Query(&vals, "playerid", w)
	if !success {
		return
	}
	iops, success := BASE.Query(&vals, "ops", w)
	if !success {
		return
	}
	data, success := BASE.Query(&vals, "data", w)
	if !success {
		return
	}

	playerid := BASE.StrToInt64(iplayerid)
	ops := BASE.StrToInt32(iops)
	
	var (
		en   int32
		rdata string
	)

	handler, exist := obj.msgJsonHandlers[ops]
	if exist {
		log.Println("required ops =", ops, " playerid=", playerid)
		filters := true
		if obj.mapFilters != nil {
			_, filters = obj.mapFilters[ops]
		}
		if filters {
			en, rdata = handler(w, r, data, playerid)
		} else {
			strToken, success := BASE.Query(&vals, "token", w)
			if !success {
				return
			}
			//需要校验token
			if int32(VP.Operation_GetIpCountryCity) == ops {
				en, rdata = handler(w, r, data, playerid)
			} else {
				if strToken == "" {
					en = int32(VP.ErrorCode_NoHandler)
					LOGGER.Error("ops not register ops= %d token is nil", ops)
					rdata = ""
				} else {
					if token.VerifyToken(playerid, strToken) == true {
						en, rdata = handler(w, r, data, playerid)
					} else {
						en = int32(VP.ErrorCode_NoHandler)
						LOGGER.Error("ops not register ops= %d token is wrong", ops)
						rdata = ""
					}
				}
			}
		}

	} else {
		en = int32(VP.ErrorCode_NoHandler)
		log.Println("ops not register ops=", ops)
		rdata = ""
	}

	LOGGER.Info("send post data:%s.", rdata)
	send := &VP.HttpJsonResult {}
	if en == int32(VP.ErrorCode_Success) {
		send.En = proto.Int32(en)
		send.Data = proto.String(rdata)
		send.Size = proto.Int32(int32(len(rdata)))
	} else {
		send.En = proto.Int32(en)
		send.Data = nil
		send.Size = proto.Int32(int32(len(rdata)))
	}
	
	res, err := json.Marshal(send)
	if err != nil {
		e := fmt.Sprintf("Mashal data error %v", err)
		LOGGER.Error(e)
	}
	
	LOGGER.Info("res post data:%s.", string(res))
	sendJsonResponse(w, 202, string(res))

	endTicks := BASE.GetCurTicks()
	if endTicks - startTicks > 1000 {
		LOGGER.Error("*msg.Ops = %d ticks = %d too long.", ops, endTicks - startTicks)
	}
}

func dispatchPost(obj *Object, w http.ResponseWriter, r *http.Request) {
	ops := int32(0)
	startTicks := BASE.GetCurTicks()
	if obj.ResponseType == Bytes {
		var (
			en   int32
			data []byte
		)
	
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sendResponse(w, 404, []byte(fmt.Sprintf(`{"res": "ioutil.ReadAll(r.Body) != err."}`)))
			LOGGER.Error("onDispatch res ioutil.ReadAll(r.Body) error:%v.", err)
			return
		}
	
		msg := &VP.Message {}
		err = proto.Unmarshal(bytes, msg)
		if err != nil {
			e := " Unmarshal failed..."
			LOGGER.Error(e)
			sendResponse(w, 404, []byte(fmt.Sprintf(`{"res": " Unmarshal failed..."}`)))
			return 
		}	
	
		ops = *msg.Ops
		handler, exist := obj.msgHandlers[*msg.Ops]
		if exist {
			log.Println("required ops =", *msg.Ops, " playerid=", *msg.PlayerId)
			filters := true
			if obj.mapFilters != nil {
				_, filters = obj.mapFilters[*msg.Ops]
			}
			if filters {
				en, data = handler(w, r, msg.Data, *msg.PlayerId)
			} else {
				//需要校验token
				strToken := *msg.Token
				if int32(VP.Operation_GetIpCountryCity) == *msg.Ops {
					en, data = handler(w, r, msg.Data, *msg.PlayerId)
				} else {
					if strToken == "" {
						en = int32(VP.ErrorCode_NoHandler)
						LOGGER.Error("ops not register ops= %d token is nil", *msg.Ops)
						data = nil
					} else {
						if token.VerifyToken(*msg.PlayerId, strToken) == true {
							en, data = handler(w, r, msg.Data, *msg.PlayerId)
						} else {
							en = int32(VP.ErrorCode_NoHandler)
							LOGGER.Error("ops not register ops= %d token is wrong", *msg.Ops)
							data = nil
						}
					}
				}
			}
	
		} else {
			en = int32(VP.ErrorCode_NoHandler)
			log.Println("ops not register ops=", *msg.Ops)
			data = nil
		}
	
		send := &VP.HttpResult {}
		if en == int32(VP.ErrorCode_Success) {
			send.En = proto.Int32(en)
			send.Data = data
			send.Size = proto.Int32(int32(len(data)))
		} else {
			send.En = proto.Int32(en)
			send.Data = nil
			send.Size = proto.Int32(int32(len(data)))
		}
		
		res, err := proto.Marshal(send)
		if err != nil {
			e := fmt.Sprintf("Mashal data error %v", err)
			LOGGER.Error(e)
		}
		sendResponse(w, 202, res)
	} else {
		var (
			en   int32
			data string
		)
	
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sendResponse(w, 404, []byte(fmt.Sprintf(`{"res": "ioutil.ReadAll(r.Body) != err."}`)))
			LOGGER.Error("onDispatch res ioutil.ReadAll(r.Body) error:%v.", err)
			return
		}
	
		msg := &VP.JsonMessage {}
		err = json.Unmarshal(bytes, msg)
		if err != nil {
			e := "json Unmarshal failed..."
			LOGGER.Error(e)
			sendJsonResponse(w, 404, fmt.Sprintf(`{"res": " Unmarshal failed..."}`))
			return 
		}	
		LOGGER.Info("get post data:%s.", string(bytes))
	
		handler, exist := obj.msgJsonHandlers[*msg.Ops]
		if exist {
			log.Println("required ops =", *msg.Ops, " playerid=", *msg.PlayerId)
			filters := true
			if obj.mapFilters != nil {
				_, filters = obj.mapFilters[*msg.Ops]
			}
			if filters {
				en, data = handler(w, r, *msg.Data, *msg.PlayerId)
			} else {
				//需要校验token
				strToken := *msg.Token
				if int32(VP.Operation_GetIpCountryCity) == *msg.Ops {
					en, data = handler(w, r, *msg.Data, *msg.PlayerId)
				} else {
					if strToken == "" {
						en = int32(VP.ErrorCode_NoHandler)
						LOGGER.Error("ops not register ops= %d token is nil", *msg.Ops)
						data = ""
					} else {
						if token.VerifyToken(*msg.PlayerId, strToken) == true {
							en, data = handler(w, r, *msg.Data, *msg.PlayerId)
						} else {
							en = int32(VP.ErrorCode_NoHandler)
							LOGGER.Error("ops not register ops= %d token is wrong", *msg.Ops)
							data = ""
						}
					}
				}
			}
	
		} else {
			en = int32(VP.ErrorCode_NoHandler)
			log.Println("ops not register ops=", *msg.Ops)
			data = ""
		}
	
		LOGGER.Info("send post data:%s.", data)
		send := &VP.HttpJsonResult {}
		if en == int32(VP.ErrorCode_Success) {
			send.En = proto.Int32(en)
			send.Data = proto.String(data)
			send.Size = proto.Int32(int32(len(data)))
		} else {
			send.En = proto.Int32(en)
			send.Data = nil
			send.Size = proto.Int32(int32(len(data)))
		}
		
		res, err := json.Marshal(send)
		if err != nil {
			e := fmt.Sprintf("Mashal data error %v", err)
			LOGGER.Error(e)
		}
		
		LOGGER.Info("res post data:%s.", string(res))
		sendJsonResponse(w, 202, string(res))
	}
	
	
	endTicks := BASE.GetCurTicks()
	if endTicks - startTicks > 1000 {
		LOGGER.Error("*msg.Ops = %d ticks = %d too long.", ops, endTicks - startTicks)
	}
}

func sendResponse(w http.ResponseWriter, code int, data []byte) {
	w.WriteHeader(code)
	w.Write(data)
}

func sendJsonResponse(w http.ResponseWriter, code int, data string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

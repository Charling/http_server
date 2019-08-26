package http_server

import (
	"net/http"
	BASE "yn.com/ext/common/function"
)

type httpHandler struct {
	ID int32
}

type Handler func(*Object, http.ResponseWriter, *http.Request)
type MsgHandler func(http.ResponseWriter, *http.Request, []byte, int64) (int32, []byte)

type Object struct {
	ID int32
	handlers map[string]Handler
	msgHandlers map[int32]MsgHandler
	ips []string
}

//go 通过nginx代理后获取用户ip
func GetIP(r *http.Request)  string{
	//nginx代理的ip
    ip := r.Header.Get("X-Real-IP")
	if ip == "" {
       ip = r.RemoteAddr
	}
	pos := BASE.UnicodeIndex(ip, ":")
	return string([]rune(ip)[:pos])
}

func CreateHttpObject(id int32) *Object {
	if mapObjects == nil {
		mapObjects = make(map[int32]*Object)
	}
	if obj,ok := mapObjects[id]; ok {
		return obj 
	}
	var str []string
	obj := &Object {
		ID: id,
		handlers: make(map[string]Handler),
		msgHandlers: make(map[int32]MsgHandler),
		ips: str,
	}
	mapObjects[id] = obj
	return obj
}

func (obj *Object) SetIps(ips []string) {
	obj.ips = nil
	obj.ips = ips
}

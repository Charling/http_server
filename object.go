package http_server

import (
	"net/http"
	BASE "github.yn.com/ext/common/function"
//	logger "github.yn.com/ext/common/logger"
	"strings"
)

const (
	Server = 0
	Client = 1
)

const (
	Bytes = 0
	Json = 1
)

type httpHandler struct {
	ID int32
}

type Handler func(*Object, http.ResponseWriter, *http.Request)
type MsgHandler func(http.ResponseWriter, *http.Request, []byte, int64) (int32, []byte)
type MsgJsonHandler func(http.ResponseWriter, *http.Request, string, int64) (int32, string)

type Object struct {
	ID int32
	ResponseType int32
	handlers map[string]Handler
	msgHandlers map[int32]MsgHandler
	msgJsonHandlers map[int32]MsgJsonHandler
	mapFilters map[int32]bool
	ips []string
}

//go 通过nginx代理后获取用户ip
func GetIP(r *http.Request)  string{
	//nginx代理的ip
	ip := r.Header.Get("X-Real-IP")
//	logger.Info("real ip:%s.", ip)
	if ip == "" {
       ip = r.RemoteAddr
	}

	if strings.ContainsAny(ip, ":") == true {
		pos := BASE.UnicodeIndex(ip, ":")
		return string([]rune(ip)[:pos])
	}
	return ip
}

func CreateHttpObject(id int32, rtype int32) *Object {
	if mapObjects == nil {
		mapObjects = make(map[int32]*Object)
	}
	if obj,ok := mapObjects[id]; ok {
		return obj 
	}
	var str []string
	obj := &Object {
		ID: id,
		ResponseType: rtype,
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

func (obj *Object) UpdateFilters(filters *map[int32]bool) {
	obj.mapFilters = *filters
}

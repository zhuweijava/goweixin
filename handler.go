package goweixin

import (
	"crypto/sha1"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var _Debug = false

const (
	TEXT     = "text"
	IMAGE    = "image"
	LOCATION = "location"
	LINK     = "link"
	EVENT    = "event"
)

type WxHttpHandler struct {
	Token   string
	Handler WxHandler
}

type WxHandler interface {
	Text(Message) Replay
	Image(Message) Replay
	Location(Message) Replay
	Link(Message) Replay
	Event(Message) Replay
	Default(Message) Replay
}

func (wx *WxHttpHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var msgType string
	var ok bool
	if _Debug {
		start := time.Now()
		defer func() {
			if ok {
				log.Printf("OK : %s , %ds", msgType, time.Now().Unix()-start.Unix())
			} else {
				log.Printf("ERR: %v", req)
			}
		}()
	}
	if err := req.ParseForm(); err != nil {
		log.Println("Bad Req", req)
		rw.WriteHeader(500)
		return
	}
	if !Verify(wx.Token, req.FormValue("timestamp"), req.FormValue("nonce"), req.FormValue("signature")) {
		rw.WriteHeader(403)
		return
	}
	if req.Method == "GET" {
		rw.Write([]byte(req.FormValue("echostr")))
		return
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("Req Body Read ERR", req, err)
		rw.WriteHeader(500)
		return
	}

	m := map[string]interface{}{}
	err = xml.Unmarshal(data, m)
	if err != nil {
		log.Println("Bad Req Body ERR", req, err)
		rw.WriteHeader(500)
		return
	}

	var msg Message
	msg = m["xml"].(map[string]interface{})
	var reply Replay
	switch msg.MsgType() {
	case TEXT:
		reply = wx.Handler.Text(msg)
	case IMAGE:
		reply = wx.Handler.Image(msg)
	case LOCATION:
		reply = wx.Handler.Location(msg)
	case LINK:
		reply = wx.Handler.Link(msg)
	case EVENT:
		reply = wx.Handler.Event(msg)
	default:
		reply = wx.Handler.Default(msg)
	}
	if reply == nil {
		return // http 200
	}

	// auto-fix
	if reply.FromUserName() == "" {
		reply.SetFromUserName(msg.ToUserName())
	}
	if reply.ToUserName() == "" {
		reply.SetToUserName(msg.FromUserName())
	}
	if reply.MsgType() == "" {
		reply.SetMsgType(TEXT)
	}
	if _, ok = reply["FuncFlag"]; !ok {
		reply.SetFuncFlag(0)
	}

	re, err := xml.Marshal(reply)
	if err != nil {
		log.Println("Bad reply", reply, err)
		rw.WriteHeader(500)
		return
	}
	rw.Write([]byte("<xml>"))
	rw.Write(re)
	rw.Write([]byte("</xml>"))
	ok = true
}

func Verify(token string, timestamp string, nonce string, signature string) bool {
	return signature == string(sha1.New().Sum([]byte(token+timestamp+nonce)))
}

type BaseWeiXinHandler struct {
}

func (h *BaseWeiXinHandler) Text(msg Message) Replay {
	return h.Default(msg)
}
func (h *BaseWeiXinHandler) Image(msg Message) Replay {
	return h.Default(msg)
}
func (h *BaseWeiXinHandler) Location(msg Message) Replay {
	return h.Default(msg)
}
func (h *BaseWeiXinHandler) Link(msg Message) Replay {
	return h.Default(msg)
}
func (h *BaseWeiXinHandler) Event(msg Message) Replay {
	return h.Default(msg)
}
func (h *BaseWeiXinHandler) Default(msg Message) Replay {
	return nil
}

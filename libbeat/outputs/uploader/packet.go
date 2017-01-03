package uploader

import (
	"fmt"
	"github.com/elastic/beats/libbeat/logp"
	"strings"
)

type Packet struct {
	Token    string
	Time	string
	HostName string
	Ip       string
	Type	string
	Message  string
	sented   *(chan bool)
}

func (p Packet) cleanMessage() string {
	s := strings.Replace(p.Message, "\n", " ", -1)
	s = strings.Replace(s, "\r", " ", -1)
	s =strings.Replace(s, "\x00", " ", -1)
	return s
}

func (p Packet) Generate() string {
	//msg := fmt.Sprintf("[token:%s time:%s host:%s ip:%s type:%s] %s",p.Token,p.Time, p.HostName,p.Ip,p.Type, p.cleanMessage())
	msg:= p.Message
	logp.Debug("log message:  %s", msg)
	return msg
}

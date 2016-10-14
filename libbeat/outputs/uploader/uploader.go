package uploader

import (
	"github.com/elastic/beats/libbeat/outputs"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/op"
	"errors"
	"github.com/elastic/beats/libbeat/logp"
)

func init() {
	outputs.RegisterOutputPlugin("uploader", New)
}

type uploaderOutput struct {
	config uploaderConfig
	logger       *Logger
}

type MapStr map[string]interface{}

func New(_ string, config *common.Config, _ int) (outputs.Outputer, error) {
	cfg := &uploaderOutput{defaultConfig,nil}
	err := config.Unpack(&cfg.config)
	ips,err:=IntranetIP()
	if(err==nil&&len(ips)>0){
		cfg.config.HostIp=ips[0]
	}
	cfg.logger, err = Dial(cfg.config.Network,cfg.config.Hosts[0],cfg.config.ConnectTimeout,cfg.config.WriteTimeout)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}


// Implement Outputer
func (out *uploaderOutput) Close() error {
	return nil
}

func (out *uploaderOutput) PublishEvent(
	sig op.Signaler,
	opts outputs.Options,
	data outputs.Data,
) error {
	event:=data.Event
	var logType string=""
	fieldsMap,ok := event["fields"].(common.MapStr)
	if ok {
		logType=common.MapStr(fieldsMap)["type"].(string)
	}
	var hostName string=""
	beatMap,ok := event["beat"].(common.MapStr)
	if ok {
		hostName=common.MapStr(beatMap)["hostname"].(string)
	}
	var timestamp string=""
	time,ok := event["@timestamp"].(common.Time)
	if ok {
		tTIme,_:=common.Time(time).MarshalJSON()
		if(len(tTIme)>2){
			timestamp=string(tTIme[1:len(tTIme)-1])
		}
	}
	var line string=""
	message,ok:=event["message"].(string)
	if(ok){
		line=message
	}

	sended := make(chan bool)
	out.logger.Packets <- Packet{
		Token:    out.config.Cid,
		Ip:       out.config.HostIp,
		HostName:	hostName,
		Type:	logType,
		Time:     timestamp,
		Message:  line,
		sented:   &sended,
	}
	if <-sended {
		op.SigCompleted(sig)
	} else {
		op.Sig(sig,errors.New("log send fail!"))
		logp.Warn("log send fail!")
	}
	return nil
}

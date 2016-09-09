package uploader

import (
	"time"
	"net"
	"github.com/elastic/beats/libbeat/logp"
	"fmt"
	"io"
)

type Logger struct {
	Packets        chan Packet
	conn           *conn
	network          string
	raddr            string
	connectTimeout   time.Duration
	writeTimeout     time.Duration
	Errors         chan error
}

type conn struct {
	netConn net.Conn
	errors  chan error
}



func Dial(network, raddr string,connectTimeout time.Duration, writeTimeout time.Duration) (*Logger, error) {
	logger := &Logger{
		Packets:          make(chan Packet, 100),
		network:          network,
		raddr:            raddr,
		connectTimeout:   connectTimeout,
		writeTimeout:     writeTimeout,
		Errors:           make(chan error, 0),
	}
	logger.connect()
	go logger.writeLoop()
	return logger, nil
}


func dial(network, raddr string,connectTimeout time.Duration) (*conn, error) {
	var netConn net.Conn
	var err error

	logp.Debug("uploader", "remote server connect timeout:%d", connectTimeout)
	netConn, err = net.DialTimeout(network, raddr, connectTimeout)

	if err != nil {
		logp.Err("connect is fail, error :%s", err.Error())
		return nil, err
	} else {
		c := &conn{netConn, make(chan error)}
		go c.watch()
		return c, nil
	}
}


// Connect to the server, retrying every 10 seconds until successful.
func (l *Logger) connect() {
	for {
		c, err := dial(l.network, l.raddr, l.connectTimeout)
		if err == nil {
			l.conn = c
			return
		} else {
			l.handleError(err)
			logp.Err("connected fail, need sleeping and connect agent,error:%s", err.Error())
			time.Sleep(10 * time.Second)
		}
	}
}


// writeloop writes any packets recieved on l.Packets() to the syslog server.
func (l *Logger) writeLoop() {
	for p := range l.Packets {
		l.writePacket(p)
	}
}

func (l *Logger) writePacket(p Packet) {
	var err error
	for {
		if l.conn.reconnectNeeded() {
			l.connect()
		}

		deadline := time.Now().Add(l.writeTimeout)
		switch l.conn.netConn.(type) {
		case *net.TCPConn,*net.UDPConn:
			l.conn.netConn.SetWriteDeadline(deadline)
			_, err = io.WriteString(l.conn.netConn, p.Generate()+"\n")
		default:
			panic(fmt.Errorf("Network protocol %s not supported", l.network))
		}
		if err == nil {
			*(p.sented) <- true
			return
		} else {
			// We had an error -- we need to close the connection and try again
			logp.Err("connect is error:%s", err.Error())
			*(p.sented) <- false
			l.conn.netConn.Close()
			l.handleError(err)
			time.Sleep(10 * time.Second)
		}
	}
}

//异步检测连接是否正常
func (c *conn) watch() {
	for {
		data := make([]byte, 1)
		_, err := c.netConn.Read(data)
		if err != nil {
			c.netConn.Close()
			c.errors <- err
			return
		}
	}
}

//重新连接
func (c *conn) reconnectNeeded() bool {
	if c == nil {
		return true
	}
	select {
	case <-c.errors:
		return true
	default:
		return false
	}
}

//处理error
func (l *Logger) handleError(err error) {
	select {
	case l.Errors <- err:
	default:
	}
}

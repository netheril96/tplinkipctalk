package lib

import (
	"bufio"
	"fmt"
	"net"
)

type TplinkTalkConnection struct {
	conn net.Conn
	rw   bufio.ReadWriter
}

func (c *TplinkTalkConnection) Start() error {
	_, err := fmt.Fprint(c.rw,
		`MULTITRANS rtsp://127.0.0.1/multitrans RTSP/1.0
CSeq: 0
Content-Length: 0
X-Handshake: unused debug
X-Client-Model: Android
X-Client-UUID: 095250a6-c01d-4af3-8ca5-7536dd45a4ff19b6d3470c5

`)
	if err != nil {
		return err
	}
	err = c.rw.Flush()
	if err != nil {
		return err
	}
	return nil
}

package lib

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
)

type TplinkTalkConnection struct {
	conn   net.Conn
	rw     bufio.ReadWriter
	user   string
	passwd string
}

type rtspResponse struct {
	StatusLine string
	Header     textproto.MIMEHeader
	Body       []byte
}

func (c *TplinkTalkConnection) Start() error {
	_, err := fmt.Fprint(c.rw, "MULTITRANS rtsp://127.0.0.1/multitrans RTSP/1.0\r\nCSeq: 0\r\nContent-Length: 0\r\nX-Handshake: unused debug\r\nX-Client-Model: Android\r\nX-Client-UUID: 095250a6-c01d-4af3-8ca5-7536dd45a4ff19b6d3470c5\r\n\r\n")
	if err != nil {
		return err
	}
	err = c.rw.Flush()
	if err != nil {
		return err
	}
	resp, err := c.readRtspResponse()
	if err != nil {
		return err
	}
	var realm, nonce string
	for _, v := range resp.Header.Values("WWW-Authenticate") {
		if strings.HasPrefix(v, "Digest") {
			realm, nonce = parseDigestAuth(v)
			break
		}
	}
	if realm == "" || nonce == "" {
		return errors.New("No realm and nonce in the WWW-Authenticate header")
	}

	uri := "rtsp://127.0.0.1/multitrans"
	method := "MULTITRANS"
	authHeader := calculateDigestAuth(c.user, c.passwd, realm, nonce, method, uri)
	fmt.Println("Authorization: " + authHeader)

	return nil
}

func calculateDigestAuth(user, passwd, realm, nonce, method, uri string) string {
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", user, realm, passwd)))
	ha2 := md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri)))
	response := md5.Sum([]byte(fmt.Sprintf("%x:%s:%x", ha1, nonce, ha2)))

	return fmt.Sprintf("Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%x\"",
		user, realm, nonce, uri, response)
}

func (c *TplinkTalkConnection) readRtspResponse() (*rtspResponse, error) {
	tp := textproto.NewReader(c.rw.Reader)
	line, err := tp.ReadLine()
	if err != nil {
		return nil, err
	}

	header, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	var body []byte
	if l := header.Get("Content-Length"); l != "" {
		length, err := strconv.Atoi(l)
		if err != nil {
			return nil, err
		}
		if length > 0 {
			body = make([]byte, length)
			if _, err := io.ReadFull(c.rw, body); err != nil {
				return nil, err
			}
		}
	}
	return &rtspResponse{
		StatusLine: line,
		Header:     header,
		Body:       body,
	}, nil
}

func parseDigestAuth(auth string) (realm string, nonce string) {
	const prefix = "Digest "
	if !strings.HasPrefix(auth, prefix) {
		return "", ""
	}
	for _, field := range strings.Split(auth[len(prefix):], ",") {
		parts := strings.SplitN(strings.TrimSpace(field), "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.Trim(parts[1], "\"")
		switch key {
		case "realm":
			realm = val
		case "nonce":
			nonce = val
		}
	}
	return
}

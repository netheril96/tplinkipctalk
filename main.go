package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/netheril96/tplinkipctalk/lib"
	"github.com/pion/rtp"
)

func parser_test() error {
	filename := `C:\Users\rsy\AppData\Local\Temp\1.g711`
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	out, err := os.Create(`C:\Users\rsy\AppData\Local\Temp\1.parsed`)
	if err != nil {
		return err
	}
	defer out.Close()
	reader := bufio.NewReader(f)
	writer := bufio.NewWriter(out)
	for {
		b, err := reader.ReadByte()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if b != '$' {
			return errors.New("Invalid packetization")
		}
		_, err = reader.ReadByte()
		if err != nil {
			return err
		}
		two_bytes := make([]byte, 2)
		_, err = io.ReadFull(reader, two_bytes)
		if err != nil {
			return err
		}
		length := (uintptr(two_bytes[0]) << 8) | uintptr(two_bytes[1])
		packet := make([]byte, length)
		_, err = io.ReadFull(reader, packet)
		if err != nil {
			return err
		}
		p := rtp.Packet{}
		err = p.Unmarshal(packet)
		if err != nil {
			return err
		}
		m, _ := json.MarshalIndent(p.Header, "", "  ")
		fmt.Println(string(m))
		_, err = writer.Write(p.Payload)
		if err != nil {
			return err
		}
	}

	return nil
}

func talk_main() error {
	user := flag.String("user", "admin", "User name")
	passwd := flag.String("passwd", "", "Password")
	ip := flag.String("ip", "10.88.40.16", "IP address")
	port := flag.String("port", "554", "Port")

	flag.Parse()
	conn, err := net.Dial("tcp", net.JoinHostPort(*ip, *port))
	if err != nil {
		return err
	}
	defer conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	tp := lib.NewTplinkTalkConnection(rw, *user, *passwd, 0)
	err = tp.Start()
	if err != nil {
		return err
	}
	defer tp.Stop()
	buf := make([]byte, 1000)
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			if err := tp.SendPcm(buf[:n]); err != nil {
				return err
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	err := talk_main()
	if err != nil {
		log.Fatal(err)
	}
}

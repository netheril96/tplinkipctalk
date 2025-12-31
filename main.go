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
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	passfile := flag.String("passfile", "", "Password file")
	ip := flag.String("ip", "10.88.40.16", "IP address")
	port := flag.String("port", "554", "Port")

	flag.Parse()
	if *passwd == "" {
		if *passfile != "" {
			data, err := os.ReadFile(*passfile)
			if err != nil {
				return err
			}
			*passwd = strings.TrimSpace(string(data))
		} else {
			return errors.New("either --passwd or --passfile is required")
		}
	}
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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	input := make(chan []byte)
	readErr := make(chan error)

	go func() {
		ticker := time.NewTicker(time.Second / 16)
		defer ticker.Stop()
		buf := make([]byte, 1000)
		for {
			<-ticker.C
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				input <- chunk
			}
			if err != nil {
				readErr <- err
				return
			}
		}
	}()

	for {
		select {
		case <-sigs:
			return nil
		case data := <-input:
			if err := tp.SendPcm(data); err != nil {
				return err
			}
		case err := <-readErr:
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

func main() {
	err := talk_main()
	if err != nil {
		log.Fatal(err)
	}
}

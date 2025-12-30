package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

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

func main() {
	err := parser_test()
	if err != nil {
		log.Fatal(err)
	}
}

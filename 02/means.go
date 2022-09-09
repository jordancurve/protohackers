// Package main implements the means to an end binary service.
// https://protohackers.com/problem/2
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var be = binary.BigEndian

func main() {
	var hostPort string
	flag.StringVar(&hostPort, "hostport", "localhost:9001", "host:port to listen on")
	flag.Parse()
	if err := doit(hostPort); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(0)
}

const readTimeout = 100 * time.Second
const writeTimeout = 100 * time.Second

type Entry struct {
	ts    int32
	price int32
}

func doit(hostPort string) error {
	addr, err := net.ResolveTCPAddr("tcp", hostPort)
	if err != nil {
		return err
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Printf("listening on %s\n", hostPort)
	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			return err
		}
		go func() {
			ph := map[int32]int32{}
			defer conn.Close()
			for {
				req := make([]byte, 9)
				conn.SetDeadline(time.Now().Add(readTimeout))
				if _, err := io.ReadFull(conn, req); err != nil {
					log.Printf("read error: %q\n", err.Error())
					return
				}
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				if _, err := conn.Write(Handle(ph, req)); err != nil {
					log.Printf("write error: %q\n", err.Error())
					return
				}
			}
		}()
	}
}

func int32ToBytes(in int32) []byte {
	var out [4]byte
	be.PutUint32(out[:], uint32(in))
	return out[:]
}

func bytesToInt32(in []byte) int32 {
	var out int32
	if err := binary.Read(bytes.NewReader(in), binary.BigEndian, &out); err != nil {
		fmt.Printf("binary.Read failed: %s", err)
	}
	return out
}

func Handle(ph map[int32]int32, req []byte) []byte {
	if len(req) != 9 {
		return nil
	}
	a, b := bytesToInt32(req[1:5]), bytesToInt32(req[5:9])
	if req[0] == 'I' {
		ph[a] = b
		return nil
	}
	if req[0] == 'Q' {
		var sum, n int64
		for k, v := range ph {
			if k >= a && k <= b {
				sum += int64(v)
				n++
			}
		}
		w := bytes.NewBuffer(nil)
		var avg int32
		if n > 0 {
			avg = int32(sum / n)
		}
		if err := binary.Write(w, binary.BigEndian, avg); err != nil {
			fmt.Printf("binary.Write failed: %s\n", err)
			return nil
		}
		return w.Bytes()
	}
	return nil
}

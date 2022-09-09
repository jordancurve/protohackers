// Package main implements the TCP Echo Service from RFC 862.
// https://protohackers.com/problem/0
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

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
		buffer := make([]byte, 1024)
		go func() {
			defer conn.Close()
			for {
				conn.SetDeadline(time.Now().Add(10 * time.Second))
				n, err := conn.Read(buffer)
				if err != nil {
					log.Println(err)
					break
				}
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if _, err := conn.Write(buffer[:n]); err != nil {
					log.Println(err)
					break
				}
			}
		}()
	}
}

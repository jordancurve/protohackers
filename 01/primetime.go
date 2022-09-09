// Package main implements the prime number JSON service
// https://protohackers.com/problem/1
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/big"
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

var isPrime = "isPrime"

const readTimeout = 100 * time.Second
const writeTimeout = 100 * time.Second

type Req struct {
	Method string   `json:"method"`
	Number *float64 `json:"number"`
}

type Resp struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
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
			defer conn.Close()
			s := bufio.NewScanner(conn)
			for {
				conn.SetDeadline(time.Now().Add(readTimeout))
				if !s.Scan() {
					break
				}
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				num, err := readReq(s.Text())
				if err != nil {
					fmt.Fprintf(conn, "{error: %q}\r\n", err.Error())
					return
				}
				resp := Resp{
					Method: "isPrime",
					Prime:  num == math.Trunc(num) && big.NewInt(int64(num)).ProbablyPrime(1),
				}
				respJson, err := json.Marshal(&resp)
				if err != nil {
					panic(err)
				}
				if _, err := conn.Write(append(respJson, '\n')); err != nil {
					log.Println(err)
					return
				}
			}
		}()
	}
}

func readReq(s string) (float64, error) {
	var m Req
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return 0, err
	}
	if m.Method != isPrime {
		return 0, fmt.Errorf("method=%q; want %q", m.Method, isPrime)
	}
	if m.Number == nil {
		return 0, fmt.Errorf("missing number")
	}
	return *m.Number, nil
}

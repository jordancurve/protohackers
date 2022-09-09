package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

type req struct {
	kind byte
	lo   int32
	hi   int32
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name string
		in   []req
		want []int32
	}{
		{
			name: "empty",
			in:   nil,
			want: nil,
		},
		{
			name: "one",
			in:   []req{{'Q', 100, 100}},
			want: []int32{0},
		},
		{
			name: "one",
			in:   []req{{'I', 100, 10}, {'Q', 100, 100}},
			want: []int32{0, 10},
		},
		{
			name: "two",
			in:   []req{{'I', 100, 10}, {'I', 110, -20}, {'Q', 100, 110}},
			want: []int32{0, 0, -5},
		},
	}
	go doit(":7890")
	time.Sleep(1 * time.Second)
	conn, err := net.Dial("tcp", ":7890")
	must(err)
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			for i := 0; i < len(test.in); i++ {
				req := encode(test.in[i])
				_, err := conn.Write(req)
				must(err)
				var buf [4]byte
				if test.in[i].kind == 'Q' {
					io.ReadFull(conn, buf[:])
				}
				got := decode(buf[:])
				if got != test.want[i] {
					t.Errorf("Handle(%v)=%v; want %v", test.in[:i+1], got, test.want[i])
				}
			}
		})
	}
}

func encode(req req) []byte {
	var enc []byte
	enc = append(enc, req.kind)
	w := bytes.NewBuffer(enc)
	binary.Write(w, binary.BigEndian, req.lo)
	binary.Write(w, binary.BigEndian, req.hi)
	return w.Bytes()
}

func decode(resp []byte) int32 {
	if len(resp) == 0 {
		return 0
	}
	var i32 int32
	if err := binary.Read(bytes.NewReader(resp), binary.BigEndian, &i32); err != nil {
		panic(err.Error())
	}

	return i32
}

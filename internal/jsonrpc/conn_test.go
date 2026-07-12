package jsonrpc

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"
)

func echoServer(conn net.Conn) {
	go func() {
		sc := bufio.NewScanner(conn)
		enc := json.NewEncoder(conn)
		for sc.Scan() {
			var req Request
			if json.Unmarshal(sc.Bytes(), &req) != nil {
				continue
			}
			if req.Method == "fail" {
				enc.Encode(Response{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32000, Message: "boom"}})
				continue
			}
			enc.Encode(Response{JSONRPC: "2.0", ID: req.ID, Result: req.Params})
		}
	}()
}

func TestCallReturnsResult(t *testing.T) {
	a, b := net.Pipe()
	echoServer(b)
	c := New(a)
	defer c.Close()
	raw, err := c.Call(context.Background(), "echo", map[string]int{"n": 7})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]int
	json.Unmarshal(raw, &got)
	if got["n"] != 7 {
		t.Fatalf("result = %v", got)
	}
}

func TestCallError(t *testing.T) {
	a, b := net.Pipe()
	echoServer(b)
	c := New(a)
	defer c.Close()
	if _, err := c.Call(context.Background(), "fail", nil); err == nil {
		t.Fatal("expected an RPC error")
	}
}

func TestCallContextCancel(t *testing.T) {
	a, b := net.Pipe()
	// A server that never responds.
	go func() { bufio.NewScanner(b).Scan() }()
	c := New(a)
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := c.Call(ctx, "slow", nil); err == nil {
		t.Fatal("expected a context deadline error")
	}
}

func TestConcurrentCalls(t *testing.T) {
	a, b := net.Pipe()
	echoServer(b)
	c := New(a)
	defer c.Close()
	done := make(chan int, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			raw, err := c.Call(context.Background(), "echo", map[string]int{"n": n})
			if err != nil {
				done <- -1
				return
			}
			var got map[string]int
			json.Unmarshal(raw, &got)
			done <- got["n"]
		}(i)
	}
	seen := map[int]bool{}
	for i := 0; i < 10; i++ {
		seen[<-done] = true
	}
	if len(seen) != 10 {
		t.Fatalf("expected 10 distinct results, got %d", len(seen))
	}
}

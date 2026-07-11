package streams

import (
	"context"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

func TestSliceStream(t *testing.T) {
	ctx := context.Background()
	s := Slice([]int{1, 2, 3})
	var got []int
	for {
		v, err := s.Next(ctx)
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Fatalf("got %v", got)
	}
	// After Close, Next ends.
	_ = s.Close()
	if _, err := s.Next(ctx); err != ports.ErrEndOfStream {
		t.Errorf("closed stream: want ErrEndOfStream, got %v", err)
	}
}

func TestSliceStreamCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Slice([]int{1}).Next(ctx); err != context.Canceled {
		t.Errorf("want context.Canceled, got %v", err)
	}
}

func TestChanStream(t *testing.T) {
	ctx := context.Background()
	st, send, closeFn := Chan[string](2)
	if !send("a") || !send("b") {
		t.Fatal("send failed")
	}
	closeFn()
	// Buffered items drain before end.
	v, err := st.Next(ctx)
	if err != nil || v != "a" {
		t.Fatalf("first = %q,%v", v, err)
	}
	// send after close is a no-op returning false
	if send("c") {
		t.Error("send after close should return false")
	}
	_ = st.Close()
}

func TestChanStreamContextCancel(t *testing.T) {
	st, _, _ := Chan[int](0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := st.Next(ctx); err != context.Canceled {
		t.Errorf("want context.Canceled, got %v", err)
	}
}

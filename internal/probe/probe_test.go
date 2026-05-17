package probe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

// freePort finds an unused TCP port on 127.0.0.1.
func freePort(t *testing.T) int {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	lis.Close()
	return port
}

// startServer starts an HTTP server on port with the given handler.
// Returns a shutdown function.
func startServer(t *testing.T, port int, h http.Handler) func() {
	t.Helper()
	srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Handler: h}
	ready := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", srv.Addr)
		if err != nil {
			t.Logf("listen %s: %v", srv.Addr, err)
			close(ready)
			return
		}
		close(ready)
		_ = srv.Serve(lis)
	}()
	<-ready
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		wg.Wait()
	}
}

func TestWaitReady_HappyPath(t *testing.T) {
	port := freePort(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- WaitReady(ctx, port)
	}()

	// Probe sees closed port → sawDown=true. Then bring up the server.
	time.Sleep(200 * time.Millisecond)
	stop := startServer(t, port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer stop()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("WaitReady: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("WaitReady did not return")
	}
}

func TestWaitReady_404IsReady(t *testing.T) {
	port := freePort(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- WaitReady(ctx, port)
	}()

	time.Sleep(200 * time.Millisecond)
	stop := startServer(t, port, http.HandlerFunc(http.NotFound))
	defer stop()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("WaitReady: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("WaitReady did not return")
	}
}

func TestWaitReady_TCPOnlyNeverReady(t *testing.T) {
	// Listener accepts but never speaks HTTP. The probe never sees the port go
	// down either (it's bound the whole test), so this should time out.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()
	go func() {
		for {
			c, err := lis.Accept()
			if err != nil {
				return
			}
			go func() {
				time.Sleep(30 * time.Second)
				c.Close()
			}()
		}
	}()

	port := lis.Addr().(*net.TCPAddr).Port
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitReady(ctx, port)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got: %v", err)
	}
}

func TestWaitReady_IPv6OnlyListener(t *testing.T) {
	// Regression: Nuxt dev server defaults to binding ::1 only on Linux.
	// The probe must reach it via "localhost" → v6 resolution.
	lis, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skipf("IPv6 loopback not available: %v", err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- WaitReady(ctx, port)
	}()

	time.Sleep(200 * time.Millisecond)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})}
	v6Lis, err := net.Listen("tcp6", fmt.Sprintf("[::1]:%d", port))
	if err != nil {
		t.Fatalf("rebind v6: %v", err)
	}
	go func() { _ = srv.Serve(v6Lis) }()
	defer srv.Shutdown(context.Background())

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("WaitReady: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("WaitReady did not return")
	}
}

func TestWaitReady_ClosedPortTimesOut(t *testing.T) {
	port := freePort(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := WaitReady(ctx, port)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded for closed port, got: %v", err)
	}
}

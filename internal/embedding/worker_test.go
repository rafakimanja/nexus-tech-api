package embedding

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerRetriesTimedOutGeneration(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		if calls.Add(1) == 1 {
			<-r.Context().Done()
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embedding": append([]float32{1}, make([]float32, 1023)...)})
	}))
	defer server.Close()
	p := &Processor{client: &http.Client{Timeout: 100 * time.Millisecond}, url: server.URL}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker := NewWorker(p)
	worker.interval = time.Millisecond
	succeeded := make(chan struct{}, 1)
	worker.process = func(ctx context.Context) error {
		_, err := p.generate(ctx, "pending demand")
		if err == nil {
			succeeded <- struct{}{}
			cancel()
		}
		return err
	}
	done := make(chan struct{})
	go func() { defer close(done); worker.Run(ctx, nil) }()
	select {
	case <-succeeded:
	case <-time.After(5 * time.Second):
		t.Fatal("pending embedding was not retried after timeout")
	}
	<-done
	if calls.Load() != 2 {
		t.Fatalf("got %d attempts, want 2", calls.Load())
	}
}

func TestWorkerSignalsDoNotBlockAndWakeNextPass(t *testing.T) {
	worker := NewWorker(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	worker.process = func(ctx context.Context) error {
		if calls.Add(1) == 1 {
			close(started)
			select {
			case <-release:
			case <-ctx.Done():
			}
		} else {
			cancel()
		}
		return nil
	}
	go func() { defer close(done); worker.Run(ctx, nil) }()
	<-started
	signaled := make(chan struct{})
	go func() {
		defer close(signaled)
		for i := uint(0); i < 1000; i++ {
			_ = worker.ProcessDemand(i)
			_ = worker.ProcessArea(i)
		}
	}()
	select {
	case <-signaled:
	case <-time.After(time.Second):
		t.Fatal("registration blocked on embedding generation")
	}
	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("new pending records did not wake the worker")
	}
}

func TestGenerateCancellation(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	p := &Processor{client: server.Client(), url: server.URL}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := p.generate(ctx, "test"); done <- err }()
	<-started
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected cancellation error")
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown did not cancel inference")
	}
}

func TestProcessorTimeoutConfiguration(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  time.Duration
	}{
		{"", 2 * time.Minute}, {"3m", 3 * time.Minute}, {"invalid", 2 * time.Minute}, {"0s", 2 * time.Minute}, {"-1s", 2 * time.Minute},
	} {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("OLLAMA_TIMEOUT", tc.value)
			if got := NewProcessor(nil).client.Timeout; got != tc.want {
				t.Fatalf("timeout = %s, want %s", got, tc.want)
			}
		})
	}
}

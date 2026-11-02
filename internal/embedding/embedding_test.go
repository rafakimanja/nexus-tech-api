package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateValidatesEmbedding(t *testing.T) {
	for _, tc := range []struct {
		name      string
		vector    []float32
		wantError bool
	}{
		{"wrong dimensions", []float32{1}, true},
		{"zero vector", make([]float32, 1024), true},
		{"valid vector", append([]float32{1}, make([]float32, 1023)...), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/embeddings" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"embedding": tc.vector})
			}))
			defer server.Close()
			p := &Processor{client: server.Client(), url: server.URL}
			_, err := p.generate(context.Background(), "test")
			if (err != nil) != tc.wantError {
				t.Fatalf("error = %v, want error %v", err, tc.wantError)
			}
		})
	}
}

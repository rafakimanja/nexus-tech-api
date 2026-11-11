package server

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloWorldHandler(t *testing.T) {
	s := &Server{}
	r := gin.New()
	r.GET("/", s.HelloWorldHandler)
	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	// Serve the HTTP request
	r.ServeHTTP(rr, req)
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// Check the response body
	expected := "{\"message\":\"Hello World\"}"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

type documentationDB struct{}

func (documentationDB) DB() *gorm.DB              { return nil }
func (documentationDB) Health() map[string]string { return map[string]string{"status": "ok"} }
func (documentationDB) Close() error              { return nil }

func TestCORSOrigins(t *testing.T) {
	for _, tc := range []struct {
		name, configured, origin string
		allowed                  bool
	}{
		{"default", "", "http://localhost:5173", true},
		{"empty list", " , , ", "http://localhost:5173", true},
		{"configured", "https://app.example.com, http://localhost:6170", "http://localhost:6170", true},
		{"other configured", "https://app.example.com, http://localhost:6170", "https://app.example.com", true},
		{"denied", "https://app.example.com", "http://localhost:5173", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CORS_ALLOWED_ORIGINS", tc.configured)
			router := (&Server{db: documentationDB{}}).RegisterRoutes()
			req := httptest.NewRequest(http.MethodOptions, "/health", nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Access-Control-Request-Method", "GET")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if got := w.Header().Get("Access-Control-Allow-Origin"); (got == tc.origin) != tc.allowed {
				t.Fatalf("origin %q: allow header %q, status %d", tc.origin, got, w.Code)
			}
		})
	}
}

func TestOpenAPIMatchesRegisteredRoutes(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	router := (&Server{db: documentationDB{}}).RegisterRoutes().(*gin.Engine)
	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(openAPISpec, &spec); err != nil {
		t.Fatal(err)
	}
	actual := make(map[string]bool)
	for _, route := range router.Routes() {
		if route.Path == "/docs" || route.Path == "/openapi.json" {
			continue
		}
		path := strings.ReplaceAll(route.Path, ":id", "{id}")
		key := route.Method + " " + path
		actual[key] = true
		if _, ok := spec.Paths[path][strings.ToLower(route.Method)]; !ok {
			t.Errorf("route missing from OpenAPI: %s", key)
		}
	}
	for path, methods := range spec.Paths {
		for method := range methods {
			key := strings.ToUpper(method) + " " + path
			if !actual[key] {
				t.Errorf("OpenAPI describes unregistered route: %s", key)
			}
		}
	}
	for _, path := range []string{"/docs", "/openapi.json"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Errorf("%s returned %d", path, w.Code)
		}
	}
}

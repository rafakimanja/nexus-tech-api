package recommendation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Set TEST_DATABASE_URL to a PostgreSQL database with the vector extension installed.
// Temporary tables isolate the test from application data.
func TestRecommendationsPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not configured")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1)
	for _, q := range []string{
		`CREATE TEMP TABLE professors (id bigint PRIMARY KEY, is_active boolean)`,
		`CREATE TEMP TABLE areas (id bigint PRIMARY KEY, name text, embedding vector(1024))`,
		`CREATE TEMP TABLE professor_areas (professor_id bigint, areas_id bigint, type text, deleted_at timestamptz)`,
		`CREATE TEMP TABLE demandas (id bigint PRIMARY KEY, title text, embedding vector(1024), professor_id bigint, is_cancelled boolean, status text, deleted_at timestamptz)`,
		`INSERT INTO professors VALUES (1, true), (2, true)`,
	} {
		if err := db.Exec(q).Error; err != nil {
			t.Fatal(err)
		}
	}
	v1 := "[1," + strings.Repeat("0,", 1022) + "0]"
	v2 := "[0,1," + strings.Repeat("0,", 1021) + "0]"
	for _, args := range [][]any{{1, "Técnica", v1}, {2, "Interesse", v2}} {
		if err := db.Exec(`INSERT INTO areas VALUES (?, ?, ?::vector)`, args...).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec(`INSERT INTO professor_areas VALUES (1, 1, 'técnica', NULL), (1, 2, 'interesse', NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]any{{1, "technical", v1, nil, false, "Em análise"}, {2, "interest", v2, nil, false, "Em análise"}, {3, "assigned", v1, 1, false, "Em análise"}, {4, "cancelled", v1, nil, true, "Em análise"}, {5, "other status", v1, nil, false, "Em desenvolvimento"}, {6, "pending", nil, nil, false, "Em análise"}} {
		if err := db.Exec(`INSERT INTO demandas VALUES (?, ?, ?::vector, ?, ?, ?, NULL)`, args...).Error; err != nil {
			t.Fatal(err)
		}
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userID", uint(1)); c.Next() })
	r.GET("/teachers/:id/recommendations", NewHandler(db).Get)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/teachers/1/recommendations", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var response struct {
		Recommendations []Recommendation `json:"recommendations"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Recommendations) != 2 || response.Recommendations[0].DemandID != 1 || response.Recommendations[1].DemandID != 2 {
		t.Fatalf("unexpected ranking: %+v", response.Recommendations)
	}
	if response.Recommendations[0].Score < 0.79 || response.Recommendations[1].Score < 0.49 {
		t.Fatalf("unexpected scores: %+v", response.Recommendations)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/teachers/2/recommendations", nil))
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", w.Code)
	}
}

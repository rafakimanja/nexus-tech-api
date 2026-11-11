package report

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestReportAuthorizationAndValidation(t *testing.T) {
	for _, tc := range []struct {
		name, role, method, path, body string
		want                           int
	}{
		{"applicant cannot report", "demandante", "POST", "/reports", `{}`, 403},
		{"professor cannot list reports", "professor", "GET", "/reports", "", 403},
		{"applicant cannot read report", "demandante", "GET", "/reports/1", "", 403},
		{"professor cannot judge", "professor", "PATCH", "/reports/1/judgment", `{}`, 403},
		{"invalid judgment id", "admin", "PATCH", "/reports/abc/judgment", `{}`, 400},
		{"invalid decision", "admin", "PATCH", "/reports/1/judgment", `{"decisao":"pendente","justificativa":"test"}`, 400},
		{"missing justification", "admin", "PATCH", "/reports/1/judgment", `{"decisao":"procedente","justificativa":" "}`, 400},
		{"malformed judgment", "admin", "PATCH", "/reports/1/judgment", `{`, 400},
		{"invalid status filter", "admin", "GET", "/reports?status=unknown", ``, 400},
		{"unknown reason", "professor", "POST", "/reports", `{"demandaId":1,"motivo":"invalid","descricao":"test"}`, 400},
		{"empty description", "professor", "POST", "/reports", `{"demandaId":1,"motivo":"spam","descricao":" "}`, 400},
		{"invalid report id", "admin", "GET", "/reports/abc", "", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) { c.Set("role", tc.role); c.Set("userID", uint(1)) })
			h := NewHandler(nil)
			r.POST("/reports", h.Create)
			r.GET("/reports", h.List)
			r.GET("/reports/:id", h.Get)
			r.PATCH("/reports/:id/judgment", h.Judge)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if w.Code != tc.want {
				t.Fatalf("got %d, want %d: %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

package demand

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func testRouter(s DemandaService, a Actor) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/demands")
	g.Use(func(c *gin.Context) { c.Set("userID", a.ID); c.Set("role", a.Role); c.Next() })
	NewRouter(NewHandler(s, nil)).RegisterRoutes(g)
	return r
}
func request(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec
}
func TestHTTPRejectsProtectedFieldsAndMalformedInputs(t *testing.T) {
	s, _, _ := fixture()
	r := testRouter(s, owner)
	for _, body := range []string{
		`{"titulo":"t","descricao":"d","publicoAlvo":"p","impacto":"i","status":"concluida"}`,
		`{"titulo":"t","descricao":"d","publicoAlvo":"p","impacto":"i","demandanteId":9}`,
		`{"titulo":"t","descricao":"d","publicoAlvo":"p","impacto":"i","professorId":2}`,
		`{"titulo":"t","descricao":"d","publicoAlvo":"p","impacto":"i","criadoEm":"2026-01-01"}`,
		`{"titulo":"t","descricao":"d","publicoAlvo":"p","impacto":"i","embedding":[1]}`,
		`{"titulo":"t","descricao":"d","publicoAlvo":"p","impacto":"i"} {}`,
		`null`, `{}`, `[]`, `{"areaId":-1}`,
	} {
		if got := request(r, "POST", "/demands/", body); got.Code != 400 {
			t.Errorf("body=%s: %d %s", body, got.Code, got.Body.String())
		}
	}
	for _, id := range []string{"0", "-1", "abc", "4294967296"} {
		if got := request(r, "GET", "/demands/"+id, ""); got.Code != 400 {
			t.Errorf("id=%s: %d", id, got.Code)
		}
	}
}
func TestHTTPDTOAndErrors(t *testing.T) {
	s, repo, _ := fixture()
	r := testRouter(s, owner)
	repo.d.Demandante.User.Name = "Nome"
	repo.d.Demandante.Type = "pessoaFisica"
	rec := request(r, "GET", "/demands/42", "")
	var dto map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || dto["demandanteNome"] != "Nome" || dto["demandanteTipo"] != "pessoaFisica" || dto["professorId"] != nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	for _, key := range []string{"Embedding", "embedding", "Demandante", "Professor", "DeletedAt", "email", "cpf_cnpj", "IsCancelled"} {
		if _, ok := dto[key]; ok {
			t.Fatalf("internal field exposed: %s", key)
		}
	}
	if got := request(r, "GET", "/demands/999", ""); got.Code != 404 {
		t.Fatal(got.Code)
	}
	repo.err = errors.New("database password=private")
	got := request(r, "GET", "/demands/42", "")
	if got.Code != 500 || strings.Contains(got.Body.String(), "private") {
		t.Fatalf("%d %s", got.Code, got.Body.String())
	}
}
func TestHTTPActionsAndDeleteCancel(t *testing.T) {
	s, repo, _ := fixture()
	r := testRouter(s, professor)
	if got := request(r, "POST", "/demands/42/analisar", ""); got.Code != 200 {
		t.Fatal(got.Code, got.Body.String())
	}
	if got := request(r, "POST", "/demands/42/assumir", `{"status":"concluida"}`); got.Code != 400 {
		t.Fatal(got.Code)
	}
	if got := request(r, "POST", "/demands/42/devolver", `{}`); got.Code != 200 {
		t.Fatal(got.Code, got.Body.String())
	}
	r = testRouter(s, owner)
	if got := request(r, "DELETE", "/demands/42", ""); got.Code != 400 {
		t.Fatal(got.Code)
	}
	if got := request(r, "DELETE", "/demands/42", `{"motivoCancelamento":"Não preciso mais"}`); got.Code != 200 {
		t.Fatal(got.Code, got.Body.String())
	}
	if repo.d.Status != Cancelada || repo.d.DeletedAt.Valid {
		t.Fatal("DELETE did not preserve cancellation")
	}
	if got := request(r, "PUT", "/demands/42", `{"titulo":"t","descricao":"d","publicoAlvo":"p","impacto":"i"}`); got.Code != 409 {
		t.Fatal(got.Code)
	}
	if got := request(testRouter(s, Actor{3, "demandante"}), "GET", "/demands/42", ""); got.Code != 403 {
		t.Fatal(got.Code)
	}
}

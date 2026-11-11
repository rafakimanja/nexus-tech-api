package report

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"api/internal/domain/demand"
	"api/internal/domain/notification"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func judgmentFixture(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not configured")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal("invalid TEST_DATABASE_URL")
	}
	base := sql.OpenDB(stdlib.GetConnector(*cfg))
	t.Cleanup(func() { base.Close() })
	schema := fmt.Sprintf("report_test_%d", time.Now().UnixNano())
	if _, err := base.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := base.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Error(err)
		}
	})
	cfg.RuntimeParams["search_path"] = schema + ",public"
	conn := sql.OpenDB(stdlib.GetConnector(*cfg))
	conn.SetMaxOpenConns(8)
	t.Cleanup(func() { conn.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`CREATE TABLE users (id bigint PRIMARY KEY, role text, deleted_at timestamptz)`,
		`CREATE TABLE demandantes (id bigint PRIMARY KEY, strikes integer NOT NULL)`,
		`CREATE TABLE professors (id bigint PRIMARY KEY, is_active boolean)`,
		`CREATE TABLE demandas (id bigint PRIMARY KEY, demandante_id bigint, professor_id bigint, status text, deleted_at timestamptz, updated_at timestamptz, is_cancelled boolean, reason_cancelled text, cancelado_em timestamptz, analise_iniciada_em timestamptz, analise_expira_em timestamptz)`,
		`INSERT INTO users VALUES (1,'admin',NULL),(2,'professor',NULL),(3,'demandante',NULL)`,
		`INSERT INTO demandantes VALUES (3,2)`,
		`INSERT INTO professors VALUES (2,true)`,
		`INSERT INTO demandas(id,demandante_id,status,is_cancelled) VALUES (1,3,'cadastrada',false),(2,3,'cadastrada',false)`,
	} {
		if err := db.Exec(q).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.AutoMigrate(&Report{}, &notification.Notification{}, &notification.Device{}, &notification.Delivery{}); err != nil {
		t.Fatal(err)
	}
	for _, r := range []Report{{ID: 1, DemandID: 1, ProfessorID: 2, Status: "pendente"}, {ID: 2, DemandID: 1, ProfessorID: 4, Status: "pendente"}, {ID: 3, DemandID: 2, ProfessorID: 2, Status: "pendente"}} {
		if err := db.Create(&r).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestJudgmentPostgres(t *testing.T) {
	db := judgmentFixture(t)
	h := NewHandler(db)
	ctx := context.Background()
	if _, err := h.judge(ctx, 1, 2, "procedente", "reason"); !errors.Is(err, demand.ErrForbidden) {
		t.Fatalf("non-admin: %v", err)
	}
	r, err := h.judge(ctx, 1, 1, "procedente", "Violação confirmada")
	if err != nil || !r.StrikeApplied || r.ReviewedBy == nil || *r.ReviewedBy != 1 || r.ReviewedAt == nil {
		t.Fatalf("judgment: %+v %v", r, err)
	}
	if _, err := h.judge(ctx, 1, 1, "improcedente", "retry"); !errors.Is(err, errAlreadyJudged) {
		t.Fatalf("retry: %v", err)
	}
	r, err = h.judge(ctx, 2, 1, "procedente", "same demand")
	if err != nil || r.StrikeApplied {
		t.Fatalf("duplicate penalty: %+v %v", r, err)
	}
	var strikes int
	if err := db.Table("demandantes").Where("id=3").Pluck("strikes", &strikes).Error; err != nil || strikes != 3 {
		t.Fatalf("strikes %d: %v", strikes, err)
	}
	var d demand.Demanda
	if err := db.First(&d, 1).Error; err != nil || d.Status != demand.Cancelada || !d.IsCancelled || d.CanceladoEm == nil {
		t.Fatalf("demand: %+v %v", d, err)
	}
	if err := demand.NewRepository(db).CheckActor(ctx, demand.Actor{ID: 3, Role: "demandante"}); !errors.Is(err, demand.ErrForbidden) {
		t.Fatalf("strike block: %v", err)
	}
	r, err = h.judge(ctx, 3, 1, "improcedente", "Sem infração")
	if err != nil || r.StrikeApplied {
		t.Fatalf("rejected: %+v %v", r, err)
	}
	var unchanged demand.Demanda
	if err := db.First(&unchanged, 2).Error; err != nil || unchanged.Status != demand.Cadastrada {
		t.Fatalf("rejection changed demand: %+v %v", unchanged, err)
	}
	var count int64
	if err := db.Model(&notification.Notification{}).Count(&count).Error; err != nil || count != 5 {
		t.Fatalf("notifications %d: %v", count, err)
	}
	for _, tc := range []struct {
		user uint
		want int
	}{{2, 200}, {9, 404}} {
		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("role", "professor"); c.Set("userID", tc.user) })
		router.GET("/reports/:id", h.Get)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", "/reports/1", nil))
		if w.Code != tc.want {
			t.Fatalf("report ownership: %d", w.Code)
		}
	}
}

func TestJudgmentConcurrentPostgres(t *testing.T) {
	db := judgmentFixture(t)
	h := NewHandler(db)
	var wg sync.WaitGroup
	for _, id := range []uint{1, 2, 3} {
		wg.Add(1)
		go func(id uint) {
			defer wg.Done()
			if _, err := h.judge(context.Background(), id, 1, "procedente", "confirmed"); err != nil {
				t.Error(err)
			}
		}(id)
	}
	wg.Wait()
	var strikes int
	if err := db.Table("demandantes").Where("id=3").Pluck("strikes", &strikes).Error; err != nil || strikes != 4 {
		t.Fatalf("concurrent strikes: %d %v", strikes, err)
	}
}

func TestReportCreationPostgres(t *testing.T) {
	db := judgmentFixture(t)
	if err := db.Exec("TRUNCATE reports RESTART IDENTITY").Error; err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("role", "professor"); c.Set("userID", uint(2)) })
	router.POST("/reports", NewHandler(db).Create)
	for _, want := range []int{201, 409} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("POST", "/reports", strings.NewReader(`{"demandaId":1,"motivo":"spam","descricao":"Spam observado"}`)))
		if w.Code != want {
			t.Fatalf("create got %d want %d: %s", w.Code, want, w.Body.String())
		}
	}
	var count int64
	if err := db.Model(&notification.Notification{}).Where("user_id=1 AND type='denunciaRecebida'").Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("admin notification: %d %v", count, err)
	}
}

func TestSameReportConcurrentPostgres(t *testing.T) {
	db := judgmentFixture(t)
	h := NewHandler(db)
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { _, err := h.judge(context.Background(), 1, 1, "procedente", "confirmed"); results <- err }()
	}
	var succeeded, conflicts int
	for i := 0; i < 2; i++ {
		err := <-results
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, errAlreadyJudged):
			conflicts++
		default:
			t.Fatal(err)
		}
	}
	if succeeded != 1 || conflicts != 1 {
		t.Fatalf("successes %d conflicts %d", succeeded, conflicts)
	}
}

func TestJudgmentRollbackPostgres(t *testing.T) {
	db := judgmentFixture(t)
	if err := db.Exec(`ALTER TABLE notifications ADD CONSTRAINT fail_delivery CHECK (type <> 'denunciaJulgada')`).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := NewHandler(db).judge(context.Background(), 1, 1, "procedente", "reason"); err == nil {
		t.Fatal("expected notification failure")
	}
	var r Report
	if err := db.First(&r, 1).Error; err != nil || r.Status != "pendente" || r.ReviewedAt != nil || r.StrikeApplied {
		t.Fatalf("report not rolled back: %+v %v", r, err)
	}
	var d demand.Demanda
	if err := db.First(&d, 1).Error; err != nil || d.Status != demand.Cadastrada {
		t.Fatalf("demand not rolled back: %+v %v", d, err)
	}
	var strikes int
	if err := db.Table("demandantes").Where("id=3").Pluck("strikes", &strikes).Error; err != nil || strikes != 2 {
		t.Fatalf("penalty not rolled back: %d %v", strikes, err)
	}
}

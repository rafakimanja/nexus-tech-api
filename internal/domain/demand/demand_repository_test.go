package demand

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// A dedicated schema keeps fixtures isolated and allows multiple connections,
// unlike TEMP tables, which cannot exercise PostgreSQL row-lock contention.
func postgresFixture(t *testing.T) *gorm.DB {
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
	schema := fmt.Sprintf("demand_test_%d", time.Now().UnixNano())
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
	// Embeddings remain NULL in these lifecycle tests; text avoids requiring
	// the vector extension for a test that only exercises relational behavior.
	for _, query := range []string{
		`CREATE TABLE users (id bigint PRIMARY KEY, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz, name text, role text, email text, foto_url text)`,
		`CREATE TABLE professors (id bigint PRIMARY KEY REFERENCES users(id), siape text, is_active boolean)`,
		`CREATE TABLE demandantes (id bigint PRIMARY KEY REFERENCES users(id), phone text, type text, cpf_cnpj text, adress text, strikes integer)`,
		`CREATE TABLE areas (id bigint PRIMARY KEY, name text)`,
		`CREATE TABLE demandas (
 id bigserial PRIMARY KEY, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz,
 demandante_id bigint NOT NULL REFERENCES demandantes(id), professor_id bigint REFERENCES professors(id),
 area_id bigint REFERENCES areas(id), title text NOT NULL, description text NOT NULL,
 publico_alvo text NOT NULL DEFAULT '', impacto text NOT NULL DEFAULT '', status text NOT NULL,
 is_cancelled boolean NOT NULL, embedding text, reason_cancelled text,
 analise_iniciada_em timestamptz, analise_expira_em timestamptz, producao_iniciada_em timestamptz,
 concluida_em timestamptz, cancelado_em timestamptz, descricao_solucao text)`,
		`INSERT INTO users(id,name,role) VALUES (1,'Demandante','demandante'),(2,'Professor A','professor'),(3,'Professor B','professor'),(4,'Outro','demandante')`,
		`INSERT INTO demandantes(id,type,strikes) VALUES (1,'pessoaFisica',0),(4,'pessoaJuridica',0)`,
		`INSERT INTO professors(id,siape,is_active) VALUES (2,'0123456',true),(3,'7654321',true)`,
		`INSERT INTO areas(id,name) VALUES (1,'Software')`,
	} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}
func TestDemandPersistencePostgres(t *testing.T) {
	db := postgresFixture(t)
	ctx := context.Background()
	repo := NewRepository(db)
	s := NewService(repo, nil)
	dto, err := s.Create(ctx, owner, validInput())
	if err != nil || dto.AreaID != nil || dto.DemandanteNome != "Demandante" || dto.DemandanteTipo != "pessoaFisica" {
		t.Fatalf("create: %+v %v", dto, err)
	}
	id := dto.ID
	area := uint(1)
	input := validInput()
	input.AreaID = &area
	dto, err = s.Update(ctx, owner, id, input)
	if err != nil || dto.AreaID == nil || *dto.AreaID != 1 {
		t.Fatalf("area: %+v %v", dto, err)
	}
	invalid := uint(999)
	input.AreaID = &invalid
	input.Titulo = "Must rollback"
	if _, err := s.Update(ctx, owner, id, input); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid area: %v", err)
	}
	dto, err = s.Get(ctx, owner, id)
	if err != nil || dto.Titulo != validInput().Titulo || *dto.AreaID != 1 {
		t.Fatalf("rollback: %+v %v", dto, err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, a := range []Actor{professor, {ID: 3, Role: "professor"}} {
		wg.Add(1)
		go func(a Actor) { defer wg.Done(); <-start; _, err := s.Act(ctx, a, id, "analisar", ""); results <- err }(a)
	}
	close(start)
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	dto, err = s.Get(ctx, owner, id)
	if err != nil || dto.ProfessorID == nil || dto.ProfessorNome == nil || dto.AnaliseExpiraEm == nil {
		t.Fatalf("reservation: %+v %v", dto, err)
	}
	assigned := Actor{ID: *dto.ProfessorID, Role: "professor"}
	if rows, err := s.List(ctx, professor, "prateleira"); err != nil || len(rows) != 0 {
		t.Fatalf("analysis on shelf: %+v %v", rows, err)
	}
	if rows, err := s.List(ctx, assigned, "minhas"); err != nil || len(rows) != 1 {
		t.Fatalf("my demands: %+v %v", rows, err)
	}

	if _, err := s.Act(ctx, assigned, id, "assumir", ""); err != nil {
		t.Fatal(err)
	}
	if rows, err := s.List(ctx, professor, "prateleira"); err != nil || len(rows) != 1 {
		t.Fatalf("production missing from shelf: %+v %v", rows, err)
	}
	if _, err := s.Act(ctx, assigned, id, "concluir", "Entregue"); err != nil {
		t.Fatal(err)
	}
	if rows, err := s.List(ctx, Actor{4, "demandante"}, ""); err != nil || len(rows) != 0 {
		t.Fatalf("other owner's data exposed: %+v %v", rows, err)
	}

	cancelled, err := s.Create(ctx, owner, validInput())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Act(ctx, owner, cancelled.ID, "cancelar", "Motivo"); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.FindByID(ctx, cancelled.ID)
	if err != nil || !stored.IsCancelled || stored.Status != Cancelada || stored.CanceladoEm == nil || stored.ReasonCancelled != "Motivo" || stored.DeletedAt.Valid {
		t.Fatalf("cancel: %+v %v", stored, err)
	}

	expired, err := s.Create(ctx, owner, validInput())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Act(ctx, professor, expired.ID, "analisar", ""); err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&Demanda{}).Where("id = ?", expired.ID).Update("analise_expira_em", time.Now().Add(-time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	restarted := NewService(NewRepository(db), nil)
	if err := restarted.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	if err := restarted.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.FindByID(ctx, expired.ID)
	if err != nil || stored.Status != Cadastrada || stored.ProfessorID != nil || stored.AnaliseExpiraEm != nil {
		t.Fatalf("expire: %+v %v", stored, err)
	}
	if _, err := restarted.Act(ctx, Actor{3, "professor"}, expired.ID, "analisar", ""); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("UPDATE professors SET is_active=false WHERE id=3").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := s.Act(ctx, Actor{3, "professor"}, expired.ID, "assumir", ""); !errors.Is(err, ErrForbidden) {
		t.Fatalf("inactive teacher: %v", err)
	}
}

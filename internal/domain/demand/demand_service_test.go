package demand

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

type memoryRepository struct {
	d            Demanda
	err          error
	createErr    error
	beforeChange func()
}

func (r *memoryRepository) CheckActor(context.Context, Actor) error { return r.err }
func (r *memoryRepository) FindByID(_ context.Context, id uint) (Demanda, error) {
	if r.err != nil {
		return Demanda{}, r.err
	}
	if r.d.ID != id {
		return Demanda{}, gorm.ErrRecordNotFound
	}
	return r.d, nil
}
func (r *memoryRepository) FindAll(context.Context, Actor, string) ([]Demanda, error) {
	return []Demanda{r.d}, r.err
}
func (r *memoryRepository) Create(_ context.Context, d *Demanda) error {
	if r.createErr != nil {
		return r.createErr
	}
	if r.err != nil {
		return r.err
	}
	d.ID = 42
	r.d = *d
	return nil
}
func (r *memoryRepository) Change(_ context.Context, id uint, fn func(*Demanda) error) (Demanda, error) {
	if r.beforeChange != nil {
		r.beforeChange()
	}
	if r.d.ID != id {
		return Demanda{}, gorm.ErrRecordNotFound
	}
	copy := r.d
	if err := fn(&copy); err != nil {
		return Demanda{}, err
	}
	r.d = copy
	return copy, nil
}
func (r *memoryRepository) Expire(_ context.Context, now time.Time) error {
	if r.err != nil {
		return r.err
	}
	if r.d.Status == EmAnalise && r.d.AnaliseExpiraEm != nil && !r.d.AnaliseExpiraEm.After(now) {
		r.d.Status = Cadastrada
		r.d.ProfessorID = nil
		r.d.AnaliseIniciadaEm = nil
		r.d.AnaliseExpiraEm = nil
		r.d.UpdatedAt = now
	}
	return nil
}

type failingEmbedder struct{ called uint }

func (e *failingEmbedder) ProcessDemand(id uint) error {
	e.called = id
	return errors.New("ollama unavailable")
}

var owner = Actor{ID: 1, Role: "demandante"}
var professor = Actor{ID: 2, Role: "professor"}

func validInput() Input {
	return Input{Titulo: "Título", Descricao: "Descrição", PublicoAlvo: "Comunidade", Impacto: "Melhoria"}
}
func fixture() (*service, *memoryRepository, *time.Time) {
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	r := &memoryRepository{d: Demanda{Model: gorm.Model{ID: 42}, DemandanteID: owner.ID, Status: Cadastrada}}
	s := NewService(r, nil).(*service)
	s.now = func() time.Time { return now }
	return s, r, &now
}
func TestCreateAndUpdateKeepDemandWhenEmbeddingFails(t *testing.T) {
	s, r, _ := fixture()
	e := &failingEmbedder{}
	s.embedder = e
	dto, err := s.Create(context.Background(), owner, validInput())
	if err != nil || dto.ID != 42 || e.called != 42 || dto.Status != Cadastrada || dto.DemandanteID != owner.ID || dto.AreaID != nil {
		t.Fatalf("dto=%+v err=%v embedding=%d", dto, err, e.called)
	}
	i := validInput()
	i.Titulo = "Changed"
	e.called = 0
	dto, err = s.Update(context.Background(), owner, 42, i)
	if err != nil || dto.Titulo != "Changed" || e.called != 42 {
		t.Fatalf("update: %+v %v", dto, err)
	}
	r.createErr = errors.New("database unavailable")
	e.called = 0
	if _, err := s.Create(context.Background(), owner, validInput()); err == nil || e.called != 0 {
		t.Fatal("failed save generated embedding")
	}
}
func TestDemandLifecycle(t *testing.T) {
	s, r, now := fixture()
	ctx := context.Background()
	dto, err := s.Act(ctx, professor, 42, "analisar", "")
	if err != nil || dto.Status != EmAnalise || dto.ProfessorID == nil || *dto.ProfessorID != 2 || !dto.AnaliseExpiraEm.Equal(now.Add(24*time.Hour)) {
		t.Fatalf("analysis: %+v %v", dto, err)
	}
	if _, err := s.Update(ctx, owner, 42, validInput()); !errors.Is(err, ErrConflict) {
		t.Fatalf("edit in analysis: %v", err)
	}
	if _, err := s.Act(ctx, owner, 42, "cancelar", "motivo"); !errors.Is(err, ErrConflict) {
		t.Fatalf("cancel in analysis: %v", err)
	}
	if _, err := s.Act(ctx, Actor{3, "professor"}, 42, "assumir", ""); !errors.Is(err, ErrForbidden) {
		t.Fatalf("other professor: %v", err)
	}
	dto, err = s.Act(ctx, professor, 42, "devolver", "")
	if err != nil || dto.Status != Cadastrada || dto.ProfessorID != nil || dto.AnaliseExpiraEm != nil {
		t.Fatalf("return: %+v %v", dto, err)
	}
	if _, err := s.Act(ctx, professor, 42, "analisar", ""); err != nil {
		t.Fatal(err)
	}
	dto, err = s.Act(ctx, professor, 42, "assumir", "")
	if err != nil || dto.Status != EmProducao || !dto.ProducaoIniciadaEm.Equal(*now) {
		t.Fatalf("production: %+v %v", dto, err)
	}
	dto, err = s.Act(ctx, professor, 42, "concluir", " Solução entregue ")
	if err != nil || dto.Status != Concluida || dto.ConcluidaEm == nil || dto.DescricaoSolucao == nil || *dto.DescricaoSolucao != "Solução entregue" {
		t.Fatalf("complete: %+v %v", dto, err)
	}
	if _, err := s.Act(ctx, professor, 42, "concluir", "outra"); !errors.Is(err, ErrConflict) {
		t.Fatal("completed demand changed")
	}
	if r.d.DeletedAt.Valid {
		t.Fatal("record deleted")
	}
}
func TestCancelPreservesRecordAndIsTerminal(t *testing.T) {
	s, r, now := fixture()
	ctx := context.Background()
	dto, err := s.Act(ctx, owner, 42, "cancelar", " Não preciso mais ")
	if err != nil || dto.Status != Cancelada || !r.d.IsCancelled || !dto.CanceladoEm.Equal(*now) || *dto.MotivoCancelamento != "Não preciso mais" {
		t.Fatalf("%+v %v", dto, err)
	}
	if _, err := s.Get(ctx, owner, 42); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(ctx, owner, 42, validInput()); !errors.Is(err, ErrConflict) {
		t.Fatal("cancelled demand edited")
	}
	if _, err := s.Act(ctx, owner, 42, "cancelar", "outro"); !errors.Is(err, ErrConflict) {
		t.Fatal("cancellation overwritten")
	}
	if r.d.DeletedAt.Valid {
		t.Fatal("cancelled record hidden")
	}
}
func TestExpirationAndRestart(t *testing.T) {
	s, r, now := fixture()
	ctx := context.Background()
	if _, err := s.Act(ctx, professor, 42, "analisar", ""); err != nil {
		t.Fatal(err)
	}
	*now = now.Add(24 * time.Hour)
	// A new service sees the persisted deadline after downtime.
	restarted := NewService(r, nil).(*service)
	restarted.now = s.now
	if err := restarted.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	if r.d.Status != Cadastrada || r.d.ProfessorID != nil || r.d.AnaliseIniciadaEm != nil || r.d.AnaliseExpiraEm != nil {
		t.Fatalf("%+v", r.d)
	}
	if err := restarted.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := restarted.Act(ctx, professor, 42, "assumir", ""); !errors.Is(err, ErrConflict) {
		t.Fatalf("expired reservation must conflict: %v", err)
	}
	if _, err := restarted.Act(ctx, Actor{3, "professor"}, 42, "analisar", ""); err != nil {
		t.Fatal(err)
	}
}
func TestDeadlineRecheckedAfterWaitingForLock(t *testing.T) {
	s, r, now := fixture()
	ctx := context.Background()
	if _, err := s.Act(ctx, professor, 42, "analisar", ""); err != nil {
		t.Fatal(err)
	}
	r.beforeChange = func() { *now = now.Add(24 * time.Hour) }
	if _, err := s.Act(ctx, professor, 42, "assumir", ""); !errors.Is(err, ErrConflict) {
		t.Fatalf("expired assignment: %v", err)
	}
	if r.d.Status != EmAnalise {
		t.Fatal("expired demand entered production")
	}
}
func TestAccessAndInputValidation(t *testing.T) {
	s, _, _ := fixture()
	ctx := context.Background()
	for _, a := range []Actor{{0, "demandante"}, {3, "demandante"}, {4, "unknown"}} {
		if _, err := s.Get(ctx, a, 42); !errors.Is(err, ErrForbidden) {
			t.Fatalf("actor=%+v err=%v", a, err)
		}
	}
	if _, err := s.Update(ctx, Actor{3, "demandante"}, 42, validInput()); !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
	if _, err := s.Create(ctx, professor, validInput()); !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
	if _, err := s.Create(ctx, owner, Input{}); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	i := validInput()
	zero := uint(0)
	i.AreaID = &zero
	if _, err := s.Create(ctx, owner, i); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.Act(ctx, owner, 42, "cancelar", " "); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, owner, 999); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}
}
func TestReadPermissions(t *testing.T) {
	for _, state := range []string{Cadastrada, EmAnalise, EmProducao, Concluida, Cancelada} {
		id := uint(2)
		d := Demanda{DemandanteID: 1, ProfessorID: &id, Status: state}
		cases := []struct {
			a       Actor
			allowed bool
		}{
			{owner, true}, {Actor{3, "demandante"}, false}, {professor, true},
			{Actor{3, "professor"}, state == Cadastrada || state == EmProducao}, {Actor{9, "admin"}, true},
		}
		for _, tc := range cases {
			if canRead(tc.a, d) != tc.allowed {
				t.Errorf("state %s actor %+v", state, tc.a)
			}
		}
	}
}

func TestEditingWithoutAreaPreservesExistingRelationship(t *testing.T) {
	s, r, _ := fixture()
	area := uint(7)
	r.d.AreaID = &area
	dto, err := s.Update(context.Background(), owner, 42, validInput())
	if err != nil || dto.AreaID == nil || *dto.AreaID != 7 {
		t.Fatalf("existing area lost: %+v %v", dto, err)
	}
}

type expiryProbe struct {
	DemandaService
	cancel context.CancelFunc
	err    error
}

func (p *expiryProbe) Expire(ctx context.Context) error {
	p.err = p.DemandaService.Expire(ctx)
	p.cancel()
	return p.err
}

func TestExpiryWorkerCatchesUpImmediatelyAndStops(t *testing.T) {
	s, r, now := fixture()
	if _, err := s.Act(context.Background(), professor, 42, "analisar", ""); err != nil {
		t.Fatal(err)
	}
	*now = now.Add(25 * time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	probe := &expiryProbe{DemandaService: s, cancel: cancel}
	RunExpiry(ctx, probe, nil)
	if probe.err != nil || r.d.Status != Cadastrada {
		t.Fatalf("startup catch-up failed: %v %s", probe.err, r.d.Status)
	}
}
func TestAllStateTransitions(t *testing.T) {
	for _, state := range []string{Cadastrada, EmAnalise, EmProducao, Concluida, Cancelada, "invalid"} {
		for _, action := range []string{"analisar", "devolver", "assumir", "concluir", "cancelar"} {
			t.Run(state+"/"+action, func(t *testing.T) {
				s, r, now := fixture()
				r.d.Status = state
				if state != Cadastrada {
					id := professor.ID
					r.d.ProfessorID = &id
				}
				expiry := now.Add(time.Hour)
				r.d.AnaliseExpiraEm = &expiry
				a := professor
				if action == "cancelar" {
					a = owner
				}
				allowed := state == Cadastrada && (action == "analisar" || action == "cancelar") ||
					state == EmAnalise && (action == "devolver" || action == "assumir") ||
					state == EmProducao && action == "concluir"
				_, err := s.Act(context.Background(), a, 42, action, "texto")
				if (err == nil) != allowed {
					t.Fatalf("allowed=%v err=%v", allowed, err)
				}
			})
		}
	}
}

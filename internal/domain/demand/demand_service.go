package demand

import (
	"context"
	"log"
	"strings"
	"time"
)

type DemandaService interface {
	List(context.Context, Actor, string) ([]DTO, error)
	Get(context.Context, Actor, uint) (DTO, error)
	Create(context.Context, Actor, Input) (DTO, error)
	Update(context.Context, Actor, uint, Input) (DTO, error)
	Act(context.Context, Actor, uint, string, string) (DTO, error)
	Expire(context.Context) error
}
type service struct {
	repo     DemandaRepository
	embedder interface{ ProcessDemand(uint) error }
	now      func() time.Time
}

func NewService(repo DemandaRepository, embedder interface{ ProcessDemand(uint) error }) DemandaService {
	return &service{repo: repo, embedder: embedder, now: func() time.Time { return time.Now().UTC() }}
}
func (s *service) prepare(ctx context.Context, a Actor) error {
	if a.ID == 0 {
		return ErrForbidden
	}
	if err := s.repo.CheckActor(ctx, a); err != nil {
		return err
	}
	return s.Expire(ctx)
}
func (s *service) Expire(ctx context.Context) error { return s.repo.Expire(ctx, s.now()) }
func (s *service) List(ctx context.Context, a Actor, scope string) ([]DTO, error) {
	if scope != "" && scope != "minhas" && scope != "prateleira" {
		return nil, ErrInvalid
	}
	if (a.Role == "demandante" && scope == "prateleira") || (a.Role == "admin" && scope != "") {
		return nil, ErrForbidden
	}
	if err := s.prepare(ctx, a); err != nil {
		return nil, err
	}
	ds, err := s.repo.FindAll(ctx, a, scope)
	if err != nil {
		return nil, err
	}
	result := make([]DTO, 0, len(ds))
	for _, d := range ds {
		result = append(result, d.DTO())
	}
	return result, nil
}
func (s *service) Get(ctx context.Context, a Actor, id uint) (DTO, error) {
	if err := s.prepare(ctx, a); err != nil {
		return DTO{}, err
	}
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return DTO{}, err
	}
	if !canRead(a, d) {
		return DTO{}, ErrForbidden
	}
	return d.DTO(), nil
}
func (s *service) Create(ctx context.Context, a Actor, i Input) (DTO, error) {
	if a.Role != "demandante" {
		return DTO{}, ErrForbidden
	}
	if err := i.validate(); err != nil {
		return DTO{}, err
	}
	if err := s.prepare(ctx, a); err != nil {
		return DTO{}, err
	}
	d := Demanda{DemandanteID: a.ID, AreaID: i.AreaID, Title: i.Titulo, Description: i.Descricao, PublicoAlvo: i.PublicoAlvo, Impacto: i.Impacto, Status: Cadastrada}
	d.CreatedAt = s.now()
	d.UpdatedAt = d.CreatedAt
	if err := s.repo.Create(ctx, &d); err != nil {
		return DTO{}, err
	}
	s.processEmbedding(d.ID)
	d, err := s.repo.FindByID(ctx, d.ID)
	return d.DTO(), err
}
func (s *service) Update(ctx context.Context, a Actor, id uint, i Input) (DTO, error) {
	if a.Role != "demandante" {
		return DTO{}, ErrForbidden
	}
	if err := i.validate(); err != nil {
		return DTO{}, err
	}
	if err := s.prepare(ctx, a); err != nil {
		return DTO{}, err
	}
	d, err := s.repo.Change(ctx, id, func(d *Demanda) error {
		if d.DemandanteID != a.ID {
			return ErrForbidden
		}
		if d.Status != Cadastrada {
			return ErrConflict
		}
		d.Title, d.Description, d.PublicoAlvo, d.Impacto = i.Titulo, i.Descricao, i.PublicoAlvo, i.Impacto
		// Older clients have no area field. Preserve any existing relationship.
		if i.AreaID != nil {
			d.AreaID = i.AreaID
		}
		d.UpdatedAt = s.now()
		return nil
	})
	if err == nil {
		s.processEmbedding(id)
	}
	return d.DTO(), err
}
func (s *service) Act(ctx context.Context, a Actor, id uint, action, text string) (DTO, error) {
	role := "professor"
	switch action {
	case "cancelar":
		role = "demandante"
	case "analisar", "devolver", "assumir", "concluir":
	default:
		return DTO{}, ErrInvalid
	}
	if a.Role != role {
		return DTO{}, ErrForbidden
	}
	text = strings.TrimSpace(text)
	if (action == "cancelar" || action == "concluir") && text == "" {
		return DTO{}, ErrInvalid
	}
	if err := s.prepare(ctx, a); err != nil {
		return DTO{}, err
	}
	d, err := s.repo.Change(ctx, id, func(d *Demanda) error {
		now := s.now()
		if action == "cancelar" && d.DemandanteID != a.ID {
			return ErrForbidden
		}
		if action != "cancelar" && action != "analisar" {
			if d.ProfessorID == nil {
				return ErrConflict
			}
			if *d.ProfessorID != a.ID {
				return ErrForbidden
			}
		}
		if d.Status == EmAnalise && (d.AnaliseExpiraEm == nil || !d.AnaliseExpiraEm.After(now)) {
			return ErrConflict
		}
		switch action {
		case "analisar":
			if d.Status != Cadastrada || d.ProfessorID != nil {
				return ErrConflict
			}
			expires := now.Add(24 * time.Hour)
			d.Status, d.ProfessorID, d.AnaliseIniciadaEm, d.AnaliseExpiraEm = EmAnalise, &a.ID, &now, &expires
		case "devolver":
			if d.Status != EmAnalise {
				return ErrConflict
			}
			d.Status, d.ProfessorID, d.AnaliseIniciadaEm, d.AnaliseExpiraEm = Cadastrada, nil, nil, nil
		case "assumir":
			if d.Status != EmAnalise {
				return ErrConflict
			}
			d.Status, d.ProducaoIniciadaEm = EmProducao, &now
		case "concluir":
			if d.Status != EmProducao {
				return ErrConflict
			}
			d.Status, d.ConcluidaEm, d.DescricaoSolucao = Concluida, &now, &text
		case "cancelar":
			if d.Status != Cadastrada {
				return ErrConflict
			}
			d.Status, d.IsCancelled, d.CanceladoEm, d.ReasonCancelled = Cancelada, true, &now, text
		}
		d.UpdatedAt = now
		return nil
	})
	return d.DTO(), err
}
func (s *service) processEmbedding(id uint) {
	if s.embedder != nil {
		if err := s.embedder.ProcessDemand(id); err != nil {
			log.Printf("demand %d embedding pending: %v", id, err)
		}
	}
}

// RunExpiry catches up on startup and then every minute.
// Reads and writes also reconcile expired reservations.
func RunExpiry(ctx context.Context, s DemandaService, logger *log.Logger) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := s.Expire(jobCtx)
		cancel()
		if err != nil && ctx.Err() == nil && logger != nil {
			logger.Printf("demand expiration failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

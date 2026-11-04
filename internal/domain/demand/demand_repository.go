package demand

import (
	"context"
	"time"

	"api/internal/domain/areas"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DemandaRepository interface {
	CheckActor(context.Context, Actor) error
	FindByID(context.Context, uint) (Demanda, error)
	FindAll(context.Context, Actor, string) ([]Demanda, error)
	Create(context.Context, *Demanda) error
	Change(context.Context, uint, func(*Demanda) error) (Demanda, error)
	Expire(context.Context, time.Time) error
}
type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) DemandaRepository { return &repository{db: db} }
func (r *repository) CheckActor(ctx context.Context, a Actor) error {
	q := r.db.WithContext(ctx).Table("users u").Where("u.id = ? AND u.role = ? AND u.deleted_at IS NULL", a.ID, a.Role)
	switch a.Role {
	case "demandante":
		q = q.Joins("JOIN demandantes p ON p.id = u.id").Where("p.strikes < 3")
	case "professor":
		q = q.Joins("JOIN professors p ON p.id = u.id").Where("p.is_active = true")
	case "admin":
	default:
		return ErrForbidden
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrForbidden
	}
	return nil
}
func details(db *gorm.DB) *gorm.DB { return db.Preload("Professor.User").Preload("Demandante.User") }
func (r *repository) FindByID(ctx context.Context, id uint) (Demanda, error) {
	var d Demanda
	err := details(r.db.WithContext(ctx)).First(&d, id).Error
	return d, err
}
func (r *repository) FindAll(ctx context.Context, a Actor, scope string) ([]Demanda, error) {
	q := details(r.db.WithContext(ctx))
	switch a.Role {
	case "demandante":
		q = q.Where("demandante_id = ?", a.ID)
	case "professor":
		if scope == "minhas" {
			q = q.Where("professor_id = ?", a.ID)
		} else {
			q = q.Where("status IN ?", []string{Cadastrada, EmProducao})
		}
	case "admin":
	default:
		return nil, ErrForbidden
	}
	ds := make([]Demanda, 0)
	err := q.Order("created_at DESC, id DESC").Find(&ds).Error
	return ds, err
}
func checkArea(db *gorm.DB, id *uint) error {
	if id == nil {
		return nil
	}
	var count int64
	if err := db.Model(&areas.Area{}).Where("id = ?", *id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrInvalid
	}
	return nil
}
func (r *repository) Create(ctx context.Context, d *Demanda) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := checkArea(tx, d.AreaID); err != nil {
			return err
		}
		return tx.Omit(clause.Associations).Create(d).Error
	})
}

// Change serializes all mutations, including competing reservations.
func (r *repository) Change(ctx context.Context, id uint, change func(*Demanda) error) (Demanda, error) {
	var d Demanda
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&d, id).Error; err != nil {
			return err
		}
		oldTitle, oldDescription, oldArea := d.Title, d.Description, d.AreaID
		if err := change(&d); err != nil {
			return err
		}
		if d.AreaID != nil && (oldArea == nil || *oldArea != *d.AreaID) {
			if err := checkArea(tx, d.AreaID); err != nil {
				return err
			}
		}
		values := map[string]any{
			"area_id": d.AreaID, "title": d.Title, "description": d.Description,
			"publico_alvo": d.PublicoAlvo, "impacto": d.Impacto, "status": d.Status,
			"professor_id": d.ProfessorID, "is_cancelled": d.Status == Cancelada,
			"reason_cancelled": d.ReasonCancelled, "cancelado_em": d.CanceladoEm,
			"analise_iniciada_em": d.AnaliseIniciadaEm, "analise_expira_em": d.AnaliseExpiraEm,
			"producao_iniciada_em": d.ProducaoIniciadaEm, "concluida_em": d.ConcluidaEm,
			"descricao_solucao": d.DescricaoSolucao, "updated_at": d.UpdatedAt,
		}
		if oldTitle != d.Title || oldDescription != d.Description {
			values["embedding"] = nil
		}
		if err := tx.Model(&d).Updates(values).Error; err != nil {
			return err
		}
		return details(tx).First(&d, id).Error
	})
	return d, err
}

// PostgreSQL rechecks the predicate after waiting for a concurrent writer.
func (r *repository) Expire(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Model(&Demanda{}).
		Where("status = ? AND analise_expira_em <= ?", EmAnalise, now).
		Updates(map[string]any{"status": Cadastrada, "professor_id": nil,
			"analise_iniciada_em": nil, "analise_expira_em": nil, "is_cancelled": false, "updated_at": now}).Error
}

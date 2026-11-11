package report

import (
	"context"
	"time"

	"api/internal/domain/demand"
	"api/internal/domain/notification"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (h *Handler) judge(ctx context.Context, id, adminID uint, decision, justification string) (Report, error) {
	var r Report
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := demand.NewRepository(tx).CheckActor(ctx, demand.Actor{ID: adminID, Role: "admin"}); err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&r, id).Error; err != nil {
			return err
		}
		if r.Status != "pendente" {
			return errAlreadyJudged
		}
		// All decisions for the same demand serialize here, including different reports.
		var d demand.Demanda
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&d, r.DemandID).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		if decision == "procedente" {
			var penalties int64
			if err := tx.Model(&Report{}).Where("demand_id = ? AND strike_applied = true", d.ID).Count(&penalties).Error; err != nil {
				return err
			}
			if penalties == 0 {
				// Atomic increment also serializes penalties on different demands of the same owner.
				result := tx.Table("demandantes").Where("id = ?", d.DemandanteID).UpdateColumn("strikes", gorm.Expr("COALESCE(strikes, 0) + 1"))
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected != 1 {
					return gorm.ErrRecordNotFound
				}
				r.StrikeApplied = true
			}
			if d.Status != demand.Cancelada {
				if err := tx.Model(&d).Updates(map[string]any{
					"status": demand.Cancelada, "is_cancelled": true,
					"reason_cancelled": justification, "cancelado_em": now,
					"analise_iniciada_em": nil, "analise_expira_em": nil, "updated_at": now,
				}).Error; err != nil {
					return err
				}
				if d.ProfessorID != nil && *d.ProfessorID != r.ProfessorID {
					if err := notification.Enqueue(tx, &notification.Notification{UserID: *d.ProfessorID, Type: "demandaCancelada", Title: "Demanda cancelada", Body: "A demanda foi cancelada após análise administrativa.", ResourceType: "demand", ResourceID: d.ID}); err != nil {
						return err
					}
				}
			}
			body := "Uma denúncia sobre sua demanda foi julgada procedente. A demanda está cancelada."
			if r.StrikeApplied {
				body += " Foi aplicado 1 strike à sua conta; 3 strikes bloqueiam o acesso."
			}
			if err := notification.Enqueue(tx, &notification.Notification{UserID: d.DemandanteID, Type: "denunciaJulgada", Title: "Resultado de denúncia", Body: body, ResourceType: "demand", ResourceID: d.ID}); err != nil {
				return err
			}
		}
		r.Status, r.Justification = decision, justification
		r.ReviewedBy, r.ReviewedAt = &adminID, &now
		if err := tx.Save(&r).Error; err != nil {
			return err
		}
		return notification.Enqueue(tx, &notification.Notification{UserID: r.ProfessorID, Type: "denunciaJulgada", Title: "Sua denúncia foi analisada", Body: "Resultado: " + decision + ". Consulte a denúncia para ver a justificativa.", ResourceType: "report", ResourceID: r.ID})
	})
	return r, err
}

package profile

import (
	"api/internal/domain/applicant"
	"api/internal/domain/areas_teacher"
	"api/internal/domain/teacher"
	"api/internal/domain/user"
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	SaveTeacher(context.Context, uint, TeacherInput, bool) error
	SaveApplicant(context.Context, uint, ApplicantInput, bool) error
}
type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func lockAccount(tx *gorm.DB, id uint, role string) error {
	var account user.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&account, id).Error; err != nil {
		return err
	}
	if account.Role != role {
		return ErrForbidden
	}
	return nil
}
func updateAccount(tx *gorm.DB, id uint, name string, photo *string) error {
	return tx.Model(&user.User{}).Where("id = ?", id).Updates(map[string]any{"name": name, "foto_url": photo}).Error
}
func (r *repository) SaveTeacher(ctx context.Context, id uint, input TeacherInput, create bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockAccount(tx, id, "professor"); err != nil {
			return err
		}
		var existing teacher.Professor
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&existing, id).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if create {
			if err == nil {
				return ErrExists
			}
			if err := tx.Omit(clause.Associations).Create(&teacher.Professor{ID: id, Siape: input.Siape, Is_active: true}).Error; err != nil {
				return err
			}
		} else {
			if err != nil {
				return err
			}
			if !existing.Is_active {
				return ErrForbidden
			}
			if err := tx.Model(&existing).Update("siape", input.Siape).Error; err != nil {
				return err
			}
		}
		if err := updateAccount(tx, id, input.Name, input.PhotoURL); err != nil {
			return err
		}
		return areas_teacher.NewRepository(tx).ReplaceByProfessorID(int(id), input.Areas)
	})
}
func (r *repository) SaveApplicant(ctx context.Context, id uint, input ApplicantInput, create bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockAccount(tx, id, "demandante"); err != nil {
			return err
		}
		var existing applicant.Demandante
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&existing, id).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if create {
			if err == nil {
				return ErrExists
			}
			if err := tx.Omit(clause.Associations).Create(&applicant.Demandante{ID: id, Phone: input.Phone, Type: input.Type, CpfCnpj: input.CpfCnpj, Adress: input.Address, Strikes: 0}).Error; err != nil {
				return err
			}
		} else {
			if err != nil {
				return err
			}
			if existing.Strikes >= 3 {
				return ErrForbidden
			}
			if err := tx.Model(&existing).Updates(map[string]any{"phone": input.Phone, "type": input.Type, "cpf_cnpj": input.CpfCnpj, "adress": input.Address}).Error; err != nil {
				return err
			}
		}
		return updateAccount(tx, id, input.Name, input.PhotoURL)
	})
}

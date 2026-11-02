package profile

import (
	"api/internal/domain/areas_teacher"
	"context"
	"strings"
)

type Service interface {
	SaveTeacher(context.Context, uint, TeacherInput, bool) error
	SaveApplicant(context.Context, uint, ApplicantInput, bool) error
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) SaveTeacher(ctx context.Context, id uint, input TeacherInput, create bool) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Siape = strings.TrimSpace(input.Siape)
	if !validCommon(input.Name, input.PhotoURL) || !ValidSiape(input.Siape) {
		return ErrInvalid
	}
	if err := areas_teacher.ValidateAssociations(input.Areas); err != nil {
		return ErrInvalid
	}
	if !HasBothAreas(input.Areas) {
		return ErrInvalid
	}
	return s.repo.SaveTeacher(ctx, id, input, create)
}
func (s *service) SaveApplicant(ctx context.Context, id uint, input ApplicantInput, create bool) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	input.CpfCnpj = strings.TrimSpace(input.CpfCnpj)
	input.Address = strings.TrimSpace(input.Address)
	if !validCommon(input.Name, input.PhotoURL) || !ValidApplicantFields(input.Phone, input.Type, input.CpfCnpj, input.Address) {
		return ErrInvalid
	}
	return s.repo.SaveApplicant(ctx, id, input, create)
}

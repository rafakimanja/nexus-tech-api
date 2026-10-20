package areas_teacher

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidAssociation = errors.New("invalid professor area association")

type AreasTeacherService interface {
	GetByProfessorID(professorID int) ([]ProfessorArea, error)
	ReplaceByProfessorID(professorID int, associations []AssociationInput) error
}

type service struct {
	repo AreasTeacherRepository
}

func NewService(repo AreasTeacherRepository) AreasTeacherService {
	return &service{repo: repo}
}

func (s *service) GetByProfessorID(professorID int) ([]ProfessorArea, error) {
	return s.repo.FindByProfessorID(professorID)
}

func (s *service) ReplaceByProfessorID(professorID int, associations []AssociationInput) error {
	seen := make(map[uint]struct{}, len(associations))
	for i := range associations {
		associations[i].Type = strings.ToLower(strings.TrimSpace(associations[i].Type))
		association := associations[i]
		if association.AreasID == 0 {
			return fmt.Errorf("%w: areas_id is required", ErrInvalidAssociation)
		}
		if _, exists := seen[association.AreasID]; exists {
			return fmt.Errorf("%w: duplicate areas_id %d", ErrInvalidAssociation, association.AreasID)
		}
		seen[association.AreasID] = struct{}{}

		switch association.Type {
		case AreaTypeTechnical, AreaTypeInterest:
		default:
			return fmt.Errorf("%w: type %q; expected %q or %q", ErrInvalidAssociation, association.Type, AreaTypeTechnical, AreaTypeInterest)
		}
	}
	return s.repo.ReplaceByProfessorID(professorID, associations)
}

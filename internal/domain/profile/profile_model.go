package profile

import (
	"api/internal/domain/areas_teacher"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var ErrInvalid = errors.New("invalid profile")
var ErrForbidden = errors.New("access denied")
var ErrExists = errors.New("profile already exists")

type TeacherInput struct {
	Name     string                           `json:"name"`
	PhotoURL *string                          `json:"photo_url"`
	Siape    string                           `json:"siape"`
	Areas    []areas_teacher.AssociationInput `json:"areas"`
}
type ApplicantInput struct {
	Name     string  `json:"name"`
	PhotoURL *string `json:"photo_url"`
	Phone    string  `json:"phone"`
	Type     string  `json:"type"`
	CpfCnpj  string  `json:"cpf_cnpj"`
	Address  string  `json:"address"`
}

func validCommon(name string, photo *string) bool {
	if strings.TrimSpace(name) == "" || len(name) > 200 {
		return false
	}
	if photo != nil {
		u, err := url.Parse(*photo)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return false
		}
	}
	return true
}
func ValidSiape(value string) bool { return regexp.MustCompile(`^[0-9]{7}$`).MatchString(value) }
func HasBothAreas(areas []areas_teacher.AssociationInput) bool {
	var technical, interest bool
	for _, a := range areas {
		technical = technical || a.Type == areas_teacher.AreaTypeTechnical
		interest = interest || a.Type == areas_teacher.AreaTypeInterest
	}
	return technical && interest
}
func ValidApplicantFields(phone, kind, document, address string) bool {
	validType := false
	for _, v := range []string{"empresa", "orgaoPublico", "ong", "instituicaoEducacional", "pessoaFisica", "academico"} {
		validType = validType || v == kind
	}
	return strings.TrimSpace(phone) != "" && validType && strings.TrimSpace(document) != "" && strings.TrimSpace(address) != ""
}

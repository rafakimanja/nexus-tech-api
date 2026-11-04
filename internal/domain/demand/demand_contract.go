package demand

import (
	"errors"
	"strings"
	"time"
)

const (
	Cadastrada = "cadastrada"
	EmAnalise  = "emAnalise"
	EmProducao = "emProducao"
	Concluida  = "concluida"
	Cancelada  = "cancelada"
)

var (
	ErrInvalid   = errors.New("invalid demand input")
	ErrForbidden = errors.New("access denied")
	ErrConflict  = errors.New("demand state conflict")
)

type Actor struct {
	ID   uint
	Role string
}

// Input excludes identity, state, dates and associations.
type Input struct {
	Titulo      string `json:"titulo"`
	Descricao   string `json:"descricao"`
	PublicoAlvo string `json:"publicoAlvo"`
	Impacto     string `json:"impacto"`
	AreaID      *uint  `json:"areaId"`
}

func (i *Input) validate() error {
	i.Titulo = strings.TrimSpace(i.Titulo)
	i.Descricao = strings.TrimSpace(i.Descricao)
	i.PublicoAlvo = strings.TrimSpace(i.PublicoAlvo)
	i.Impacto = strings.TrimSpace(i.Impacto)
	if i.Titulo == "" || i.Descricao == "" || i.PublicoAlvo == "" || i.Impacto == "" || (i.AreaID != nil && *i.AreaID == 0) {
		return ErrInvalid
	}
	return nil
}

type DTO struct {
	ID                 uint       `json:"id"`
	DemandanteID       uint       `json:"demandanteId"`
	DemandanteNome     string     `json:"demandanteNome"`
	DemandanteTipo     string     `json:"demandanteTipo"`
	Titulo             string     `json:"titulo"`
	Descricao          string     `json:"descricao"`
	PublicoAlvo        string     `json:"publicoAlvo"`
	Impacto            string     `json:"impacto"`
	AreaID             *uint      `json:"areaId"`
	Status             string     `json:"status"`
	MotivoCancelamento *string    `json:"motivoCancelamento"`
	ProfessorID        *uint      `json:"professorId"`
	ProfessorNome      *string    `json:"professorNome"`
	AnaliseIniciadaEm  *time.Time `json:"analiseIniciadaEm"`
	AnaliseExpiraEm    *time.Time `json:"analiseExpiraEm"`
	ProducaoIniciadaEm *time.Time `json:"producaoIniciadaEm"`
	ConcluidaEm        *time.Time `json:"concluidaEm"`
	DescricaoSolucao   *string    `json:"descricaoSolucao"`
	CriadoEm           time.Time  `json:"criadoEm"`
	AtualizadoEm       time.Time  `json:"atualizadoEm"`
	CanceladoEm        *time.Time `json:"canceladoEm"`
}

func (d Demanda) DTO() DTO {
	var nome, motivo *string
	if d.ProfessorID != nil {
		v := d.Professor.User.Name
		nome = &v
	}
	if d.ReasonCancelled != "" {
		v := d.ReasonCancelled
		motivo = &v
	}
	return DTO{ID: d.ID, DemandanteID: d.DemandanteID, DemandanteNome: d.Demandante.User.Name,
		DemandanteTipo: d.Demandante.Type, Titulo: d.Title, Descricao: d.Description,
		PublicoAlvo: d.PublicoAlvo, Impacto: d.Impacto, AreaID: d.AreaID, Status: d.Status,
		MotivoCancelamento: motivo, ProfessorID: d.ProfessorID, ProfessorNome: nome,
		AnaliseIniciadaEm: d.AnaliseIniciadaEm, AnaliseExpiraEm: d.AnaliseExpiraEm,
		ProducaoIniciadaEm: d.ProducaoIniciadaEm, ConcluidaEm: d.ConcluidaEm,
		DescricaoSolucao: d.DescricaoSolucao, CriadoEm: d.CreatedAt, AtualizadoEm: d.UpdatedAt, CanceladoEm: d.CanceladoEm}
}
func canRead(a Actor, d Demanda) bool {
	return a.Role == "admin" || (a.Role == "demandante" && d.DemandanteID == a.ID) ||
		(a.Role == "professor" && (d.Status == Cadastrada || d.Status == EmProducao || (d.ProfessorID != nil && *d.ProfessorID == a.ID)))
}

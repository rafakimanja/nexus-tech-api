package report

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api/internal/domain/demand"
	"api/internal/domain/notification"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Report preserves the original complaint and the administrator's final decision.
type Report struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	DemandID      uint       `json:"demandaId" gorm:"not null;uniqueIndex:report_professor_demand"`
	ProfessorID   uint       `json:"professorId" gorm:"not null;uniqueIndex:report_professor_demand"`
	Reason        string     `json:"motivo" gorm:"not null"`
	Description   string     `json:"descricao" gorm:"not null"`
	Status        string     `json:"status" gorm:"not null"`
	CreatedAt     time.Time  `json:"criadoEm"`
	ReviewedBy    *uint      `json:"julgadoPor"`
	ReviewedAt    *time.Time `json:"julgadoEm"`
	Justification string     `json:"justificativa"`
	StrikeApplied bool       `json:"strikeAplicado" gorm:"not null;default:false"`
}
type Handler struct{ db *gorm.DB }

func NewHandler(db *gorm.DB) *Handler { return &Handler{db: db} }
func fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
func (h *Handler) Create(c *gin.Context) {
	if c.GetString("role") != "professor" {
		fail(c, 403, "access denied")
		return
	}
	var input struct {
		DemandID    uint   `json:"demandaId"`
		Reason      string `json:"motivo"`
		Description string `json:"descricao"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "invalid report")
		return
	}
	input.Description = strings.TrimSpace(input.Description)
	reasons := map[string]bool{"conteudoOfensivo": true, "spam": true, "foraDoEscopo": true, "informacaoFalsa": true, "demandaDuplicada": true, "outro": true}
	if input.DemandID == 0 || !reasons[input.Reason] || input.Description == "" || len(input.Description) > 5000 {
		fail(c, 400, "invalid report")
		return
	}
	a := demand.Actor{ID: c.GetUint("userID"), Role: c.GetString("role")}
	r := Report{DemandID: input.DemandID, ProfessorID: a.ID, Reason: input.Reason, Description: input.Description, Status: "pendente"}
	err := h.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		repo := demand.NewRepository(tx)
		if err := repo.CheckActor(c.Request.Context(), a); err != nil {
			return err
		}
		d, err := repo.FindByID(c.Request.Context(), input.DemandID)
		if err != nil {
			return err
		}
		if d.Status != demand.Cadastrada && d.Status != demand.EmProducao && (d.ProfessorID == nil || *d.ProfessorID != a.ID) {
			return demand.ErrForbidden
		}
		if err := tx.Create(&r).Error; err != nil {
			return err
		}
		var admins []uint
		if err := tx.Table("users").Where("role = 'admin' AND deleted_at IS NULL").Pluck("id", &admins).Error; err != nil {
			return err
		}
		for _, id := range admins {
			if err := notification.Enqueue(tx, &notification.Notification{UserID: id, Type: "denunciaRecebida", Title: "Nova denúncia recebida", Body: "Uma demanda foi denunciada. Consulte os detalhes na área administrativa.", ResourceType: "report", ResourceID: r.ID}); err != nil {
				return err
			}
		}
		return nil
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		fail(c, 409, "demand already reported by this professor")
		return
	}
	if errors.Is(err, demand.ErrForbidden) {
		fail(c, 403, "access denied")
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fail(c, 404, "demand not found")
		return
	}
	if err != nil {
		fail(c, 500, "could not create report")
		return
	}
	c.JSON(201, r)
}
func (h *Handler) List(c *gin.Context) {
	if c.GetString("role") != "admin" {
		fail(c, 403, "access denied")
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "30"))
	if err != nil || limit < 1 || limit > 100 {
		fail(c, 400, "invalid limit")
		return
	}
	status := c.Query("status")
	if status != "" {
		if status != "pendente" && status != "procedente" && status != "improcedente" {
			fail(c, 400, "invalid report status")
			return
		}
	}
	q := h.db.WithContext(c.Request.Context())
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if raw := c.Query("before"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			fail(c, 400, "invalid cursor")
			return
		}
		q = q.Where("id < ?", id)
	}
	items := make([]Report, 0)
	if err := q.Order("id DESC").Limit(limit).Find(&items).Error; err != nil {
		fail(c, 500, "could not retrieve reports")
		return
	}
	c.JSON(200, items)
}
func (h *Handler) Get(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "professor" {
		fail(c, 403, "access denied")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		fail(c, 400, "invalid report ID")
		return
	}
	q := h.db.WithContext(c.Request.Context())
	if role == "professor" {
		q = q.Where("professor_id = ?", c.GetUint("userID"))
	}
	var r Report
	err = q.First(&r, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fail(c, 404, "report not found")
		return
	}
	if err != nil {
		fail(c, 500, "could not retrieve report")
		return
	}
	c.JSON(200, r)
}

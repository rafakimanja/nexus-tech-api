package recommendation

import (
	"database/sql"
	_ "embed"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//go:embed suggest_demands.sql
var suggestDemandsSQL string

type Area struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Recommendation struct {
	DemandID    uint    `json:"demand_id"`
	Title       string  `json:"title"`
	Score       float64 `json:"score"`
	MatchedArea Area    `json:"matched_area"`
}

type row struct {
	DemandID uint
	Title    string
	Score    float64
	AreaID   uint
	AreaName string
	AreaType string
}

type Handler struct{ db *gorm.DB }

func NewHandler(db *gorm.DB) *Handler { return &Handler{db: db} }

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		respondError(c, http.StatusBadRequest, "invalid professor ID")
		return
	}
	userID, ok := c.Get("userID")
	if !ok || userID != uint(id) {
		respondError(c, http.StatusForbidden, "access denied")
		return
	}
	limit := 10
	if value := c.Query("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 100 {
			respondError(c, http.StatusBadRequest, "limit must be between 1 and 100")
			return
		}
	}
	var professor struct {
		ID       uint
		IsActive bool `gorm:"column:is_active"`
	}
	if err := h.db.WithContext(c.Request.Context()).Table("professors").Select("id, is_active").Where("id = ?", id).Take(&professor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "professor not found")
		} else {
			respondError(c, http.StatusInternalServerError, "could not retrieve professor")
		}
		return
	}
	recommendations := make([]Recommendation, 0)
	if professor.IsActive {
		var rows []row
		err = h.db.WithContext(c.Request.Context()).Raw(
			suggestDemandsSQL,
			sql.Named("professor_id", id),
			sql.Named("limit", limit),
		).Scan(&rows).Error
		if err != nil {
			respondError(c, http.StatusInternalServerError, "could not retrieve recommendations")
			return
		}
		for _, r := range rows {
			recommendations = append(recommendations, Recommendation{DemandID: r.DemandID, Title: r.Title, Score: r.Score, MatchedArea: Area{ID: r.AreaID, Name: r.AreaName, Type: r.AreaType}})
		}
	}
	c.JSON(http.StatusOK, gin.H{"teacher_id": id, "recommendations": recommendations})
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}

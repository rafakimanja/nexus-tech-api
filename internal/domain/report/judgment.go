package report

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"api/internal/domain/demand"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var errAlreadyJudged = errors.New("report already judged")

// Judge accepts one final decision. Repeated submissions never apply another penalty.
func (h *Handler) Judge(c *gin.Context) {
	if c.GetString("role") != "admin" {
		fail(c, http.StatusForbidden, "access denied")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		fail(c, 400, "invalid report ID")
		return
	}
	var input struct {
		Decision      string `json:"decisao"`
		Justification string `json:"justificativa"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "invalid judgment")
		return
	}
	input.Justification = strings.TrimSpace(input.Justification)
	if (input.Decision != "procedente" && input.Decision != "improcedente") || input.Justification == "" || len(input.Justification) > 5000 {
		fail(c, 400, "invalid judgment")
		return
	}
	r, err := h.judge(c.Request.Context(), uint(id), c.GetUint("userID"), input.Decision, input.Justification)
	switch {
	case errors.Is(err, errAlreadyJudged):
		fail(c, 409, "report already judged")
	case errors.Is(err, demand.ErrForbidden):
		fail(c, 403, "access denied")
	case errors.Is(err, gorm.ErrRecordNotFound):
		fail(c, 404, "report or demand not found")
	case err != nil:
		fail(c, 500, "could not judge report")
	default:
		c.JSON(200, r)
	}
}

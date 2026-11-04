package demand

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DemandaHandler struct {
	service DemandaService
	logger  *log.Logger
}

func NewHandler(s DemandaService, logger *log.Logger) *DemandaHandler {
	return &DemandaHandler{service: s, logger: logger}
}
func actor(c *gin.Context) Actor { return Actor{ID: c.GetUint("userID"), Role: c.GetString("role")} }
func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		return 0, ErrInvalid
	}
	return uint(id), nil
}
func decode(c *gin.Context, input any, allowEmpty bool) error {
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(input); err != nil {
		if allowEmpty && errors.Is(err, io.EOF) {
			return nil
		}
		return ErrInvalid
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ErrInvalid
	}
	return nil
}
func (h *DemandaHandler) GetDemandas(c *gin.Context) {
	result, err := h.service.List(c.Request.Context(), actor(c), c.Query("escopo"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(200, result)
}
func (h *DemandaHandler) GetDemandaByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		h.fail(c, err)
		return
	}
	result, err := h.service.Get(c.Request.Context(), actor(c), id)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(200, result)
}
func (h *DemandaHandler) CreateDemanda(c *gin.Context) {
	var input Input
	if err := decode(c, &input, false); err != nil {
		h.fail(c, err)
		return
	}
	result, err := h.service.Create(c.Request.Context(), actor(c), input)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(201, result)
}
func (h *DemandaHandler) UpdateDemanda(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		h.fail(c, err)
		return
	}
	var input Input
	if err := decode(c, &input, false); err != nil {
		h.fail(c, err)
		return
	}
	result, err := h.service.Update(c.Request.Context(), actor(c), id, input)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(200, result)
}
func (h *DemandaHandler) Action(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseID(c)
		if err != nil {
			h.fail(c, err)
			return
		}
		var text string
		switch action {
		case "cancelar":
			var input struct {
				Motivo string `json:"motivoCancelamento"`
			}
			err = decode(c, &input, false)
			text = input.Motivo
		case "concluir":
			var input struct {
				Solucao string `json:"descricaoSolucao"`
			}
			err = decode(c, &input, false)
			text = input.Solucao
		default:
			var input struct{}
			err = decode(c, &input, true)
		}
		if err != nil {
			h.fail(c, err)
			return
		}
		result, err := h.service.Act(c.Request.Context(), actor(c), id, action, text)
		if err != nil {
			h.fail(c, err)
			return
		}
		c.JSON(200, result)
	}
}
func (h *DemandaHandler) fail(c *gin.Context, err error) {
	status, message := 500, "could not process demand"
	switch {
	case errors.Is(err, ErrInvalid):
		status, message = 400, "invalid demand fields, area or ID"
	case errors.Is(err, ErrForbidden):
		status, message = 403, "access denied"
	case errors.Is(err, gorm.ErrRecordNotFound):
		status, message = 404, "demand not found"
	case errors.Is(err, ErrConflict):
		status, message = 409, "action not allowed in the current demand state"
	default:
		if h.logger != nil {
			h.logger.Printf("demand operation failed: %v", err)
		}
	}
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}

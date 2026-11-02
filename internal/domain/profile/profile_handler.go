package profile

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	service Service
	logger  *log.Logger
}

func NewHandler(service Service, logger *log.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) SaveTeacher(c *gin.Context) {
	id, create, ok := owner(c, "professor")
	if !ok {
		return
	}
	var input TeacherInput
	if err := decode(c, &input); err != nil {
		respond(c, 400, "invalid profile or unsupported fields")
		return
	}
	if err := h.service.SaveTeacher(c.Request.Context(), id, input, create); err != nil {
		h.fail(c, err)
		return
	}
	if create {
		c.JSON(201, gin.H{"teacherID": id})
	} else {
		c.Status(204)
	}
}
func (h *Handler) SaveApplicant(c *gin.Context) {
	id, create, ok := owner(c, "demandante")
	if !ok {
		return
	}
	var input ApplicantInput
	if err := decode(c, &input); err != nil {
		respond(c, 400, "invalid profile or unsupported fields")
		return
	}
	if err := h.service.SaveApplicant(c.Request.Context(), id, input, create); err != nil {
		h.fail(c, err)
		return
	}
	if create {
		c.JSON(201, gin.H{"applicantID": id})
	} else {
		c.Status(204)
	}
}
func owner(c *gin.Context, role string) (uint, bool, bool) {
	id := c.GetUint("userID")
	create := c.Request.Method == http.MethodPost
	if id == 0 || c.GetString("role") != role {
		respond(c, 403, "access denied")
		return 0, create, false
	}
	if !create {
		pathID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil || pathID == 0 {
			respond(c, 400, "invalid ID")
			return 0, create, false
		}
		if uint(pathID) != id {
			respond(c, 403, "access denied")
			return 0, create, false
		}
	}
	return id, create, true
}
func decode(c *gin.Context, input any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(input); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ErrInvalid
	}
	return nil
}
func (h *Handler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		respond(c, 400, "invalid profile fields")
	case errors.Is(err, ErrForbidden):
		respond(c, 403, "access denied")
	case errors.Is(err, ErrExists):
		respond(c, 409, "profile already exists; use PUT to edit")
	case errors.Is(err, gorm.ErrRecordNotFound):
		respond(c, 404, "profile, account or area not found")
	default:
		if h.logger != nil {
			h.logger.Printf("profile save failed: %v", err)
		}
		respond(c, 500, "could not save profile")
	}
}
func respond(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}

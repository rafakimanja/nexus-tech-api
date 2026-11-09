package notification

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Handler struct { db *gorm.DB }
func NewHandler(db *gorm.DB) *Handler { return &Handler{db:db} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("", h.List)
	r.GET("/unread-count", h.Count)
	r.PATCH("/read-all", h.ReadAll)
	r.PATCH("/:id/read", h.Read)
	r.PUT("/devices", h.RegisterDevice)
	r.DELETE("/devices", h.RemoveDevice)
}
func fail(c *gin.Context, status int, message string) { c.JSON(status, gin.H{"error":gin.H{"code":status,"message":message}}) }
func (h *Handler) owned(c *gin.Context) *gorm.DB { return h.db.WithContext(c.Request.Context()).Model(&Notification{}).Where("user_id = ?", c.GetUint("userID")) }
func (h *Handler) List(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "30")); if err != nil || limit < 1 || limit > 100 { fail(c,400,"limit must be between 1 and 100"); return }
	q := h.owned(c)
	if raw := c.Query("before"); raw != "" { id, err := strconv.ParseUint(raw,10,64); if err != nil || id == 0 { fail(c,400,"invalid cursor"); return }; q = q.Where("id < ?",id) }
	items := make([]Notification,0)
	if err := q.Order("id DESC").Limit(limit).Find(&items).Error; err != nil { fail(c,500,"could not retrieve notifications"); return }
	c.JSON(200,items)
}
func (h *Handler) Count(c *gin.Context) { var count int64; if err:=h.owned(c).Where("read_at IS NULL").Count(&count).Error; err != nil { fail(c,500,"could not count notifications"); return }; c.JSON(200,gin.H{"count":count}) }
func (h *Handler) Read(c *gin.Context) {
	id,err:=strconv.ParseUint(c.Param("id"),10,64); if err!=nil || id==0 { fail(c,400,"invalid notification ID"); return }
	result:=h.owned(c).Where("id = ?",id).Update("read_at",gorm.Expr("COALESCE(read_at, ?)",time.Now().UTC()))
	if result.Error!=nil { fail(c,500,"could not mark notification as read"); return }; if result.RowsAffected==0 { fail(c,404,"notification not found"); return }; c.Status(204)
}
func (h *Handler) ReadAll(c *gin.Context) { if err:=h.owned(c).Where("read_at IS NULL").Update("read_at",time.Now().UTC()).Error; err!=nil { fail(c,500,"could not mark notifications as read"); return }; c.Status(204) }
func tokenInput(c *gin.Context) (string,bool) {
	c.Request.Body=http.MaxBytesReader(c.Writer,c.Request.Body,4096)
	var body struct { Token string `json:"token"` }
	if c.ShouldBindJSON(&body)!=nil || len(body.Token)<1 || len(body.Token)>2048 || strings.ContainsAny(body.Token," \t\r\n") { fail(c,400,"invalid device token"); return "",false }; return body.Token,true
}
func (h *Handler) RegisterDevice(c *gin.Context) {
	token,ok:=tokenInput(c); if !ok{return}
	d:=Device{UserID:c.GetUint("userID"),Token:token}
	err:=h.db.WithContext(c.Request.Context()).Clauses(clause.OnConflict{Columns:[]clause.Column{{Name:"token"}},DoUpdates:clause.Assignments(map[string]any{"user_id":d.UserID,"updated_at":time.Now().UTC()})}).Create(&d).Error
	if err!=nil { fail(c,500,"could not register device"); return }; c.Status(204)
}
func (h *Handler) RemoveDevice(c *gin.Context) {
	token,ok:=tokenInput(c); if !ok{return}
	if err:=h.db.WithContext(c.Request.Context()).Where("user_id = ? AND token = ?",c.GetUint("userID"),token).Delete(&Device{}).Error; err!=nil { fail(c,500,"could not remove device"); return }; c.Status(204)
}

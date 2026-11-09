package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrUnregistered = errors.New("device is no longer registered")
type Sender interface { Send(context.Context, Device, Notification) error }
type FCMSender struct { client *http.Client; endpoint string }

// Credentials are supplied through Application Default Credentials on the VPS.
// They must never be shipped in the mobile application.
func NewFCMSender(ctx context.Context, project string) (*FCMSender,error) {
	client,err:=google.DefaultClient(ctx,"https://www.googleapis.com/auth/firebase.messaging")
	if err!=nil{return nil,err}; client.Timeout=10*time.Second
	return &FCMSender{client:client,endpoint:"https://fcm.googleapis.com/v1/projects/"+url.PathEscape(project)+"/messages:send"},nil
}
func (s *FCMSender) Send(ctx context.Context, device Device, n Notification) error {
	payload:=map[string]any{"message":map[string]any{
		"token":device.Token,"notification":map[string]string{"title":n.Title,"body":n.Body},
		"data":map[string]string{"notification_id":strconv.FormatUint(uint64(n.ID),10),"resource_type":n.ResourceType,"resource_id":strconv.FormatUint(uint64(n.ResourceID),10)},
		"android":map[string]any{"priority":"HIGH","notification":map[string]string{"tag":"notification-"+strconv.FormatUint(uint64(n.ID),10)}},
	}}
	body,err:=json.Marshal(payload); if err!=nil{return err}
	req,err:=http.NewRequestWithContext(ctx,http.MethodPost,s.endpoint,bytes.NewReader(body)); if err!=nil{return err}; req.Header.Set("Content-Type","application/json")
	resp,err:=s.client.Do(req); if err!=nil{return err}; defer resp.Body.Close()
	if resp.StatusCode>=200 && resp.StatusCode<300 {return nil}
	var result struct { Error struct { Details []struct { ErrorCode string `json:"errorCode"` } `json:"details"` } `json:"error"` }
	_ = json.NewDecoder(io.LimitReader(resp.Body,64<<10)).Decode(&result)
	for _,d:=range result.Error.Details { if d.ErrorCode=="UNREGISTERED" {return ErrUnregistered} }
	// Never include the response body or token in logs.
	return fmt.Errorf("FCM returned HTTP %d",resp.StatusCode)
}

// Dispatch uses row locks to support concurrent API processes. Network sending
// is at-least-once: a crash after FCM accepts a message can cause a retry.
func Dispatch(ctx context.Context,db *gorm.DB,sender Sender) error {
	for i:=0;i<20;i++ {
		err:=db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var delivery Delivery
			if err:=tx.Clauses(clause.Locking{Strength:"UPDATE",Options:"SKIP LOCKED"}).Where("finished_at IS NULL AND next_attempt_at <= ?",time.Now().UTC()).Order("id").First(&delivery).Error; err!=nil{return err}
			var n Notification; if err:=tx.First(&n,delivery.NotificationID).Error; err!=nil{return err}
			var device Device; err:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).First(&device,delivery.DeviceID).Error
			now:=time.Now().UTC()
			if errors.Is(err,gorm.ErrRecordNotFound) || (err==nil && device.UserID!=n.UserID) {return tx.Model(&delivery).Update("finished_at",now).Error}; if err!=nil{return err}
			err=sender.Send(ctx,device,n)
			if errors.Is(err,ErrUnregistered) {if err:=tx.Delete(&device).Error; err!=nil{return err}; err=nil}
			delivery.Attempts++
			if err==nil || delivery.Attempts>=8 { delivery.FinishedAt=&now } else {delivery.NextAttemptAt=now.Add(time.Duration(1<<delivery.Attempts)*30*time.Second)}
			return tx.Save(&delivery).Error
		})
		if errors.Is(err,gorm.ErrRecordNotFound){return nil}; if err!=nil{return err}
	}
	return nil
}
func RunPush(ctx context.Context,db *gorm.DB,sender Sender,logger *log.Logger) {
	ticker:=time.NewTicker(5*time.Second); defer ticker.Stop()
	for { job,cancel:=context.WithTimeout(ctx,25*time.Second); err:=Dispatch(job,db,sender); cancel(); if err!=nil && ctx.Err()==nil {logger.Printf("notification dispatch failed: %v",err)}; select {case <-ctx.Done():return; case <-ticker.C:} }
}

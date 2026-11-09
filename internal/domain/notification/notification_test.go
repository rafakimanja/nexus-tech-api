package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFCMSender(t *testing.T) {
	for _, tc := range []struct { name string; status int; response string; want error }{
		{"success",200,`{"name":"projects/test/messages/1"}`,nil},
		{"expired token",404,`{"error":{"details":[{"errorCode":"UNREGISTERED"}]}}`,ErrUnregistered},
		{"permission failure",403,`{"error":{"message":"private detail"}}`,errors.New("failure")},
	} {
		t.Run(tc.name,func(t *testing.T){
			server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
				if r.Method!="POST" || r.Header.Get("Content-Type")!="application/json" {t.Error("invalid request")}
				var body struct { Message struct { Token string; Notification map[string]string; Data map[string]string } }
				if err:=json.NewDecoder(r.Body).Decode(&body);err!=nil{t.Error(err)}
				if body.Message.Token!="device-token" || body.Message.Data["resource_id"]!="42" || body.Message.Notification["title"]!="Updated" {t.Errorf("invalid payload: %+v",body)}
				w.WriteHeader(tc.status); _,_=w.Write([]byte(tc.response))
			}));defer server.Close()
			sender:=FCMSender{client:server.Client(),endpoint:server.URL}
			err:=sender.Send(context.Background(),Device{Token:"device-token"},Notification{ID:1,Title:"Updated",ResourceType:"demand",ResourceID:42})
			if tc.want==nil && err!=nil{t.Fatal(err)};if tc.want!=nil && err==nil{t.Fatal("expected error")};if tc.want==ErrUnregistered && !errors.Is(err,ErrUnregistered){t.Fatal(err)}
			if err!=nil && strings.Contains(err.Error(),"private detail"){t.Fatal("response leaked")}
		})
	}
}

func fixture(t *testing.T) *gorm.DB {
	t.Helper(); dsn:=os.Getenv("TEST_DATABASE_URL");if dsn==""{t.Skip("TEST_DATABASE_URL not configured")}
	db,err:=gorm.Open(postgres.Open(dsn),&gorm.Config{});if err!=nil{t.Fatal(err)}
	conn,err:=db.DB();if err!=nil{t.Fatal(err)};conn.SetMaxOpenConns(1);t.Cleanup(func(){conn.Close()})
	for _,q:=range []string{
		`CREATE TEMP TABLE notifications (id bigserial PRIMARY KEY,user_id bigint NOT NULL,type text,title text,body text,resource_type text,resource_id bigint,event_key text UNIQUE,created_at timestamptz,read_at timestamptz)`,
		`CREATE TEMP TABLE devices (id bigserial PRIMARY KEY,user_id bigint,token text UNIQUE,created_at timestamptz,updated_at timestamptz)`,
		`CREATE TEMP TABLE deliveries (id bigserial PRIMARY KEY,notification_id bigint,device_id bigint,attempts integer,next_attempt_at timestamptz,finished_at timestamptz,UNIQUE(notification_id,device_id))`,
	}{if err:=db.Exec(q).Error;err!=nil{t.Fatal(err)}}
	return db
}

type fakeSender struct { calls int; err error }
func (s *fakeSender) Send(_ context.Context,_ Device,_ Notification)error{s.calls++;return s.err}

func TestDispatchIdleQueuePostgres(t *testing.T) {
	db := fixture(t)
	var logs bytes.Buffer
	db = db.Session(&gorm.Session{Logger: logger.New(log.New(&logs, "", 0), logger.Config{LogLevel: logger.Warn})})
	sender := &fakeSender{}
	ctx := context.Background()
	if err := Dispatch(ctx, db, sender); err != nil {
		t.Fatalf("empty queue: %v", err)
	}
	if logs.Len() != 0 || sender.calls != 0 {
		t.Fatalf("empty queue logged an error or sent a message: calls=%d logs=%s", sender.calls, logs.String())
	}
	// A delivery scheduled for later also means there is no work right now.
	delivery := Delivery{NotificationID: 999, DeviceID: 999, NextAttemptAt: time.Now().Add(time.Hour)}
	if err := db.Create(&delivery).Error; err != nil {
		t.Fatal(err)
	}
	logs.Reset()
	if err := Dispatch(ctx, db, sender); err != nil {
		t.Fatalf("future delivery: %v", err)
	}
	if logs.Len() != 0 || sender.calls != 0 {
		t.Fatalf("idle queue logged an error or sent a message: calls=%d logs=%s", sender.calls, logs.String())
	}
	// Missing related data must not be mistaken for an idle queue.
	if err := db.Model(&delivery).Update("next_attempt_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	if err := Dispatch(ctx, db, sender); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing notification must propagate: %v", err)
	}
}

func TestInboxAndDeliveryPostgres(t *testing.T) {
	db:=fixture(t);ctx:=context.Background()
	device:=Device{UserID:1,Token:"old-token"};if err:=db.Create(&device).Error;err!=nil{t.Fatal(err)}
	key:="event-1"
	for i:=0;i<2;i++{if err:=db.Transaction(func(tx *gorm.DB)error{return Enqueue(tx,&Notification{UserID:1,Type:"test",EventKey:&key})});err!=nil{t.Fatal(err)}}
	var count int64;db.Model(&Notification{}).Count(&count);if count!=1{t.Fatalf("duplicate notification: %d",count)}
	db.Model(&Delivery{}).Count(&count);if count!=1{t.Fatalf("duplicate delivery: %d",count)}
	sentinel:=errors.New("rollback")
	_ = db.Transaction(func(tx *gorm.DB)error{if err:=Enqueue(tx,&Notification{UserID:1,Type:"rollback"});err!=nil{t.Fatal(err)};return sentinel})
	db.Model(&Notification{}).Count(&count);if count!=1{t.Fatal("notification escaped rollback")}
	h:=NewHandler(db);r:=gin.New();r.Use(func(c *gin.Context){c.Set("userID",uint(2))});h.Register(r.Group("/notifications"))
	w:=httptest.NewRecorder();r.ServeHTTP(w,httptest.NewRequest("PATCH","/notifications/1/read",nil));if w.Code!=404{t.Fatalf("read other user's notification: %d",w.Code)}
	w=httptest.NewRecorder();r.ServeHTTP(w,httptest.NewRequest("GET","/notifications",nil));if w.Code!=200 || strings.TrimSpace(w.Body.String())!="[]"{t.Fatalf("inbox leaked: %s",w.Body.String())}
	sender:=&fakeSender{err:errors.New("temporary failure")};if err:=Dispatch(ctx,db,sender);err!=nil{t.Fatal(err)}
	var d Delivery;db.First(&d);if d.Attempts!=1 || d.FinishedAt!=nil || !d.NextAttemptAt.After(time.Now()){t.Fatalf("retry not scheduled: %+v",d)}
	// The same phone changes accounts before the retry. Do not send the old user's message.
	db.Model(&device).Update("user_id",2);db.Model(&d).Update("next_attempt_at",time.Now().Add(-time.Minute))
	if err:=Dispatch(ctx,db,sender);err!=nil{t.Fatal(err)};if sender.calls!=1{t.Fatal("sent to device belonging to another account")}
}

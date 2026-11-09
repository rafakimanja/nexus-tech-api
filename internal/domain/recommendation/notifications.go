package recommendation

import (
	"context"
	"fmt"
	"log"
	"time"

	"api/internal/domain/notification"
	"gorm.io/gorm"
)

// NotifyNew uses the same ranking as the recommendation endpoint: the top ten
// available demands per active professor. Each professor/demand pair is notified
// once, even after restarts, profile changes or a demand returning to the shelf.
func NotifyNew(ctx context.Context, db *gorm.DB) error {
	var rows []struct { ProfessorID uint; DemandID uint; Title string }
	err:=db.WithContext(ctx).Raw(`WITH scores AS (
		SELECT p.id AS professor_id, d.id AS demand_id, d.title,
		MAX(GREATEST(0, 1 - (d.embedding <=> a.embedding)) *
		CASE pa.type WHEN 'técnica' THEN 0.8 WHEN 'interesse' THEN 0.5 END) AS score
		FROM professors p JOIN users u ON u.id=p.id AND u.deleted_at IS NULL AND u.role='professor'
		JOIN professor_areas pa ON pa.professor_id=p.id AND pa.deleted_at IS NULL
		JOIN areas a ON a.id=pa.areas_id AND a.embedding IS NOT NULL
		CROSS JOIN demandas d
		WHERE p.is_active=true AND d.deleted_at IS NULL AND d.professor_id IS NULL
		AND d.is_cancelled=false AND lower(trim(d.status))='cadastrada'
		AND d.embedding IS NOT NULL AND pa.type IN ('técnica','interesse')
		GROUP BY p.id,d.id,d.title
	), ranked AS (
		SELECT *,row_number() OVER (PARTITION BY professor_id ORDER BY score DESC,demand_id) AS rank FROM scores
	) SELECT professor_id,demand_id,title FROM ranked WHERE rank<=10
	AND NOT EXISTS (SELECT 1 FROM notifications n WHERE n.event_key='recommendation:' || professor_id || ':' || demand_id)
	ORDER BY professor_id,demand_id LIMIT 500`).Scan(&rows).Error
	if err!=nil{return err}
	for _,r:=range rows {
		key:=fmt.Sprintf("recommendation:%d:%d",r.ProfessorID,r.DemandID)
		if err:=db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {return notification.Enqueue(tx,&notification.Notification{UserID:r.ProfessorID,Type:"demandaRecomendada",Title:"Demanda recomendada para você",Body:r.Title,ResourceType:"demand",ResourceID:r.DemandID,EventKey:&key})}); err!=nil{return err}
	}
	return nil
}

func RunNotifications(ctx context.Context,db *gorm.DB,logger *log.Logger) {
	ticker:=time.NewTicker(30*time.Second); defer ticker.Stop()
	for {job,cancel:=context.WithTimeout(ctx,25*time.Second); err:=NotifyNew(job,db); cancel(); if err!=nil && ctx.Err()==nil {logger.Printf("recommendation notifications failed: %v",err)}; select {case <-ctx.Done():return; case <-ticker.C:} }
}

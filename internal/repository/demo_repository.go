package repository

import (
	"context"

	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
	"strings"
	"time"
)

type DemoRepository struct {
	db *gorm.DB
}

func NewDemoRepository(db *gorm.DB) *DemoRepository {
	return &DemoRepository{db: db}
}

type DemoResetStats struct {
	ReservationsCancelled int64
	ReservationsDeleted   int64
	ZonesDeleted          int64
	AuditDeleted          int64
}

func (r *DemoRepository) ResetSession(ctx context.Context, sessionID string) (DemoResetStats, error) {
	var stats DemoResetStats
	if sessionID == "" {
		return stats, nil
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var activeIDs []uint
		if err := tx.Model(&models.Reservation{}).
			Where("demo_session_id = ? AND status = ?", sessionID, models.ReservationStatusActive).
			Pluck("id", &activeIDs).Error; err != nil {
			return err
		}
		if len(activeIDs) > 0 {
			res := tx.Model(&models.Reservation{}).
				Where("id IN ?", activeIDs).
				Updates(map[string]any{"status": models.ReservationStatusCancelled})
			if res.Error != nil {
				return res.Error
			}
			stats.ReservationsCancelled = res.RowsAffected
		}

		res := tx.Where("demo_session_id = ?", sessionID).Delete(&models.Reservation{})
		if res.Error != nil {
			return res.Error
		}
		stats.ReservationsDeleted = res.RowsAffected

		res = tx.Where("demo_session_id = ? AND is_demo = false", sessionID).Delete(&models.ParkingZone{})
		if res.Error != nil {
			return res.Error
		}
		stats.ZonesDeleted = res.RowsAffected

		res = tx.Where("demo_session_id = ?", sessionID).Delete(&models.AuditLog{})
		if res.Error != nil {
			return res.Error
		}
		stats.AuditDeleted = res.RowsAffected
		return nil
	})
	return stats, err
}

func (r *DemoRepository) ListStaleSessionIDs(ctx context.Context, inactiveBefore time.Time) ([]string, error) {
	type row struct {
		DemoSessionID string
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT demo_session_id
		FROM reservations
		WHERE demo_session_id IS NOT NULL AND demo_session_id <> ''
		GROUP BY demo_session_id
		HAVING MAX(created_at) < ?
		   AND SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) = 0
	`, inactiveBefore, models.ReservationStatusActive).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.DemoSessionID) != "" {
			out = append(out, strings.TrimSpace(row.DemoSessionID))
		}
	}
	return out, nil
}

func (r *DemoRepository) PurgeStaleSessions(ctx context.Context, inactiveBefore time.Time) (int, DemoResetStats, error) {
	ids, err := r.ListStaleSessionIDs(ctx, inactiveBefore)
	if err != nil {
		return 0, DemoResetStats{}, err
	}
	var total DemoResetStats
	for _, id := range ids {
		stats, err := r.ResetSession(ctx, id)
		if err != nil {
			return len(ids), total, err
		}
		total.ReservationsCancelled += stats.ReservationsCancelled
		total.ReservationsDeleted += stats.ReservationsDeleted
		total.ZonesDeleted += stats.ZonesDeleted
		total.AuditDeleted += stats.AuditDeleted
	}
	return len(ids), total, nil
}

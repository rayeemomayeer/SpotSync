package repository

import (
	"context"

	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
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

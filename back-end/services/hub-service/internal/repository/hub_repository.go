package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/outbox"
	"github.com/nusaroute/services/hub-service/internal/model"
)

type HubRepository interface {
	GetHubByID(ctx context.Context, id string) (*model.Hub, error)
	ListHubs(ctx context.Context) ([]model.Hub, error)
	CreateScanLog(ctx context.Context, scan *model.ScanLog, outboxTopic string, outboxPayload interface{}) error
	GetScansByAWB(ctx context.Context, awb string) ([]model.ScanLog, error)
	GetManifest(ctx context.Context, hubID string, date time.Time) ([]model.ScanLog, error)
	GetDashboardStats(ctx context.Context) (activeHubs int64, totalCities int64, err error)
}

type hubRepo struct{ db *sqlx.DB }

func NewHubRepository(db *sqlx.DB) HubRepository { return &hubRepo{db: db} }

func (r *hubRepo) GetHubByID(ctx context.Context, id string) (*model.Hub, error) {
	var h model.Hub
	err := r.db.GetContext(ctx, &h, "SELECT * FROM hubs WHERE id = $1", id)
	return &h, err
}

func (r *hubRepo) ListHubs(ctx context.Context) ([]model.Hub, error) {
	var hubs []model.Hub
	err := r.db.SelectContext(ctx, &hubs, "SELECT * FROM hubs WHERE is_active = true ORDER BY name")
	return hubs, err
}

func (r *hubRepo) CreateScanLog(ctx context.Context, scan *model.ScanLog, outboxTopic string, outboxPayload interface{}) error {
	scan.ID = uuid.New().String()
	scan.ScannedAt = time.Now()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO scan_logs (id, awb, order_id, hub_id, scan_type, operator_id, note, scanned_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err = tx.ExecContext(ctx, query,
		scan.ID, scan.AWB, scan.OrderID, scan.HubID, scan.ScanType, scan.OperatorID, scan.Note, scan.ScannedAt)
	if err != nil {
		return err
	}

	if outboxTopic != "" {
		if err := outbox.InsertEvent(ctx, tx, outboxTopic, outboxPayload); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *hubRepo) GetScansByAWB(ctx context.Context, awb string) ([]model.ScanLog, error) {
	var scans []model.ScanLog
	err := r.db.SelectContext(ctx, &scans, "SELECT * FROM scan_logs WHERE awb = $1 ORDER BY scanned_at", awb)
	return scans, err
}

func (r *hubRepo) GetManifest(ctx context.Context, hubID string, date time.Time) ([]model.ScanLog, error) {
	var scans []model.ScanLog
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	err := r.db.SelectContext(ctx, &scans,
		"SELECT * FROM scan_logs WHERE hub_id = $1 AND scanned_at >= $2 AND scanned_at < $3 ORDER BY scanned_at",
		hubID, startOfDay, endOfDay)
	return scans, err
}

func (r *hubRepo) GetDashboardStats(ctx context.Context) (activeHubs int64, totalCities int64, err error) {
	err = r.db.GetContext(ctx, &activeHubs, "SELECT COUNT(*) FROM hubs WHERE is_active = true")
	if err != nil { return 0, 0, err }

	err = r.db.GetContext(ctx, &totalCities, "SELECT COUNT(DISTINCT city) FROM hubs")
	return activeHubs, totalCities, err
}

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/services/hub-service/internal/model"
)

type HubRepository interface {
	GetHubByID(ctx context.Context, id string) (*model.Hub, error)
	ListHubs(ctx context.Context) ([]model.Hub, error)
	CreateScanLog(ctx context.Context, scan *model.ScanLog) error
	GetScansByAWB(ctx context.Context, awb string) ([]model.ScanLog, error)
	GetManifest(ctx context.Context, hubID string, date time.Time) ([]model.ScanLog, error)
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

func (r *hubRepo) CreateScanLog(ctx context.Context, scan *model.ScanLog) error {
	scan.ID = uuid.New().String()
	scan.ScannedAt = time.Now()
	query := `INSERT INTO scan_logs (id, awb, order_id, hub_id, scan_type, operator_id, note, scanned_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.db.ExecContext(ctx, query,
		scan.ID, scan.AWB, scan.OrderID, scan.HubID, scan.ScanType, scan.OperatorID, scan.Note, scan.ScannedAt)
	return err
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
